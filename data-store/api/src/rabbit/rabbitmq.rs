extern crate lapin_futures as lapin;
extern crate futures;
extern crate tokio_core;
extern crate pretty_env_logger;
extern crate serde_json;

use std::{panic, time, thread, env, io};
use std::sync::{Arc, mpsc};
use std::sync::mpsc::Receiver;
use self::futures::future::Future;
use self::futures::Stream;
use self::tokio_core::reactor::Core;
use self::tokio_core::net::TcpStream;
use self::lapin::types::FieldTable;
use self::lapin::client::ConnectionOptions;
use self::lapin::channel::{
    BasicProperties,
    BasicPublishOptions,
    BasicConsumeOptions,
    ExchangeDeclareOptions,
    QueueDeclareOptions};

use db::database::DB;

#[derive(Deserialize)]
struct Payload {
    device: String,
    sensor: String,
    rssi: i32,
    time: u64,
}

fn event_exchange(rx: Receiver<(Payload, String)>, rabbit_host: String) {
    let exchange_name = "event";

    // create the reactor
    let mut core = Core::new().unwrap();
    let handle = core.handle();
    let addr = rabbit_host.parse().unwrap();

    core.run(

        TcpStream::connect(&addr, &handle).and_then(|stream| {

            // connect() returns a future of an AMQP Client
            // that resolves once the handshake is done
            lapin::client::Client::connect(stream, &ConnectionOptions::default())
        }).and_then(|(client, heartbeat_future_fn)| {

            let heartbeat_client = client.clone();
            thread::Builder::new().name("exch heartbeat thread".to_string()).spawn(move || {
                Core::new().unwrap().run(heartbeat_future_fn(&heartbeat_client)).unwrap();
            }).unwrap();

            println!("Exch connected!");
            // create channel to rabbit
            client.create_channel().and_then(move |channel| {
                // Set Exchange options
                let mut exch_options = ExchangeDeclareOptions::default();
                exch_options.durable = true;

                // declare topic exchange
                channel.exchange_declare(exchange_name, "topic", &exch_options, &FieldTable::new()).and_then(move |_| {

                    // extract and construct routing key from JSON payload
                    let extract_routing_key = |payload: Payload| {
                        format!(
                            "SENSOR.{}.DETECTED.{}",
                            payload.sensor.clone(),
                            payload.device.clone())
                    };
                    loop {

                        debug!("Waiting for event...");
                        let (json, data) = rx.recv()
                            .expect("channel message failed");

                        let ch = channel.clone();
                        let routing_key = extract_routing_key(json);

                        // publish to event exchange
                        ch.basic_publish(
                            exchange_name,
                            &routing_key,
                            data.as_bytes(),
                            &BasicPublishOptions::default(),
                            BasicProperties::default()
                        ).wait()
                         .expect("Couldn't publish event!");

                    }
                    Ok(())

                })
            })
        })
    ).unwrap();
}

pub fn run(rabbit_host: String) {

    // Connect to DB
    let db_addr = env::var("DB_HOST_ADDR")
        .expect("Database Host Address must be provided through env var!");

    info!("Rabbit consumer connecting to DB at http://{}", db_addr);
    let db = Arc::new(DB::new(&format!("http://{}", db_addr)));

    // create the reactor
    let mut core = Core::new().unwrap();
    let handle = core.handle();
    let addr = rabbit_host.parse().unwrap();

    debug!("Running Rabbit driver towards {}", addr);

    let (tx, rx) = mpsc::channel();

    thread::Builder::new().name("exchange thread".to_string()).spawn(move || {
        debug!("Starting exch thread!");
        event_exchange(rx, rabbit_host);
    }).unwrap();

    core.run(

        TcpStream::connect(&addr, &handle).and_then(|stream| {

            // connect() returns a future of an AMQP Client
            // that resolves once the handshake is done
            lapin::client::Client::connect(stream, &ConnectionOptions::default())
        }).and_then(|(client, heartbeat_future_fn)| {

            // The heartbeat future should be run in a dedicated thread so that nothing can prevent it from
            // dispatching events on time.
            // If we ran it as part of the "main" chain of futures, we might end up not sending
            // some heartbeats if we don't poll often enough (because of some blocking task or such).
            let heartbeat_client = client.clone();
            thread::Builder::new().name("heartbeat thread".to_string()).spawn(move || {
                Core::new().unwrap().run(heartbeat_future_fn(&heartbeat_client)).unwrap();
            }).unwrap();

            // create_channel returns a future that is resolved
            // once the channel is successfully created
            client.create_channel()
        }).and_then(|channel| {
            let id = channel.id;
            debug!("created channel with id: {}", id);

            let queue = "sensor_data";
            let ch = channel.clone();

            channel.queue_declare(queue, &QueueDeclareOptions::default(), &FieldTable::new()).and_then(move |_| {
                info!("channel {} declared queue {}", id, queue);

                // basic_consume returns a future of a message
                // stream. Any time a message arrives for this consumer,
                // the for_each method would be called
                channel.basic_consume(queue, "data_store", &BasicConsumeOptions::default(), &FieldTable::new())
            }).and_then(|stream| {
                info!("got consumer stream");

                stream.for_each(move |message| {
                    let data = String::from_utf8(message.data).unwrap();
                    info!("got message: {}", data);

                    let json: Payload = serde_json::from_str(&data).unwrap();
                    db.insert(json.device.clone(), json.sensor.clone(), json.rssi.clone(), json.time.clone());
                    ch.basic_ack(message.delivery_tag);
                    tx.send((json, data))
                        .expect("Couldn't send to exch thread!");

                    Ok(())
                })
            })
        })
    ).unwrap();

}

