extern crate lapin_futures as lapin;
extern crate futures;
extern crate tokio_core;
extern crate env_logger;

use std::{self, thread};
use self::futures::future::Future;
use self::futures::Stream;
use self::tokio_core::reactor::Core;
use self::tokio_core::net::TcpStream;
use self::lapin::types::FieldTable;
use self::lapin::client::ConnectionOptions;
use self::lapin::channel::{BasicConsumeOptions, QueueDeclareOptions};

pub fn run(rabbit_host: String) {
  env_logger::init().unwrap();

  // create the reactor
  let mut core = Core::new().unwrap();
  let handle = core.handle();
  let addr = rabbit_host.parse().unwrap(); //"127.0.0.1:5672"

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
		  info!("created channel with id: {}", id);

		  let ch = channel.clone();
		  channel.queue_declare("hello", &QueueDeclareOptions::default(), &FieldTable::new()).and_then(move |_| {
			info!("channel {} declared queue {}", id, "hello");

			// basic_consume returns a future of a message
			// stream. Any time a message arrives for this consumer,
			// the for_each method would be called
			channel.basic_consume("hello", "my_consumer", &BasicConsumeOptions::default(), &FieldTable::new())
		  }).and_then(|stream| {
			info!("got consumer stream");

			stream.for_each(move |message| {
			  debug!("got message: {:?}", message);
			  info!("decoded message: {:?}", std::str::from_utf8(&message.data).unwrap());
			  ch.basic_ack(message.delivery_tag);
			  Ok(())
			})
		  })
		})
	  ).unwrap();

}

