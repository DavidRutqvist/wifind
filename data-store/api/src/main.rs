#![feature(plugin, decl_macro, type_ascription)]
#![plugin(rocket_codegen)]

#[macro_use] extern crate influx_db_client;
#[macro_use] extern crate log;
#[macro_use] extern crate serde_derive;
#[macro_use] extern crate serde_json;
extern crate serde;
extern crate rocket;
extern crate rocket_contrib;

use std::{time, env};
use std::thread;

use rocket::State;
use rocket::response::status::NotFound;
use rocket_contrib::Json;

mod db;
use db::database::DB;
mod consul;
use consul::sd;
mod rabbit;
use rabbit::rabbitmq;

#[derive(Deserialize)]
struct Payload {
    device: String,
    sensor: String,
    rssi: i32,
    time: u64,
}

#[post("/", data = "<data>")]
fn post_device(data: Json<Payload>, db: State<DB>) -> &'static str {

    db.insert(data.device.clone(), data.sensor.clone(), data.rssi.clone(), data.time.clone());
    return "success"

}

#[get("/<start>/<stop>")]
fn time_interval(start: u64, stop: u64, db: State<DB>) -> Json<serde_json::Value> {

    let results = db.get_time_interval(start, stop);
    Json(results)
}

#[get("/<sensor_id>")]
fn sensor(sensor_id: String, db: State<DB>) -> Result<Json<serde_json::Value>, NotFound<Json<serde_json::Value>>> {
    let result = db.get_sensor(sensor_id.clone());

    let res: Vec<serde_json::Value> = serde_json::from_value(result.clone()).unwrap();
    match res.len() {
        0 => Err(NotFound(Json(json!({"success": false, "message": format!("Sensor {} not found", sensor_id)})))),
        _ => Ok(Json(result))
    }
}

#[get("/")]
fn sensors(db: State<DB>) -> Json<serde_json::Value> {
    let result = db.get_sensors();
    Json(result)
}

#[get("/<device_hash>")]
fn device(device_hash: String, db: State<DB>) -> Json<serde_json::Value> {
    let result = db.get_device(device_hash);
    Json(result)
}

#[get("/")]
fn dump(db: State<DB>) -> Json<serde_json::Value> {

    let result = db.dump();
    Json(result)
}

#[get("/")]
fn health_check(db: State<DB>) -> Result<&'static str, &'static str> {
    match db.is_responding() {
        true => Ok("Healthy!"),
        false => Err("DB non responsive!")
    }
}

#[get("/")]
fn root() -> Json<serde_json::Value> {
    Json(json!({"success": true, "message": "This is an api"}))
}

fn main() {

    let rabbit_addr = sd::get_node_address("rabbit")
        .expect("No RabbitMQ service found! Is Consul reachable?");

    thread::spawn(move || {
        loop {
            let addr = rabbit_addr.clone();
            if thread::Builder::new().name("rabbit thread".to_string()).spawn(move || {
                rabbitmq::run(addr);
            }).unwrap().join().is_err() {
                warn!("Connection to Rabbit failed, restarting in 1 second...");
                thread::sleep(time::Duration::from_millis(1000));
            }
        }
    });

    let db_addr = env::var("DB_HOST_ADDR")
        .expect("Database Host Address must be provided through env var!");
    let db = DB::new(&format!("http://{}", db_addr));

    rocket::ignite()
        .manage(db)
        .mount("/", routes![root, post_device])
        .mount("/health", routes![health_check])
        .mount("/dump", routes![dump])
        .mount("/time", routes![time_interval])
        .mount("/sensor", routes![sensor, sensors])
        .mount("/device", routes![device])
        .launch();
}

