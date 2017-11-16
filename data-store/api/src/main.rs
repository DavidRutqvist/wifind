#![feature(plugin, decl_macro, type_ascription)]
#![plugin(rocket_codegen)]

#[macro_use]
extern crate serde_derive;
#[macro_use]
extern crate serde_json;
extern crate serde;
#[macro_use]
extern crate influx_db_client;
extern crate rocket;
extern crate rocket_contrib;

use std::env;

use rocket::State;
use rocket_contrib::Json;

mod db;
use db::database::DB;

#[derive(Deserialize)]
struct Payload {
    device: String,
    sensor: String,
    rssi: i32,
    timestamp: u64,
}

#[post("/", data = "<data>")]
fn post_device(data: Json<Payload>, db: State<DB>) -> &'static str {

    db.insert(data.device.clone(), data.sensor.clone(), data.rssi.clone(), data.timestamp.clone());
    return "success"

}

#[get("/<start>/<stop>")]
fn time_interval(start: u64, stop: u64, db: State<DB>) -> Json<serde_json::Value> {

    let results = db.get_time_interval(start, stop);
    Json(results)
}

#[get("/<sensor_id>")]
fn sensor(sensor_id: String, db: State<DB>) -> Json<serde_json::Value> {
    let result = db.get_sensor(sensor_id);
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
fn root() -> &'static str {
    "This is an API"
}

fn main() {

    let db_addr = env::var("DB_HOST_ADDR")
        .expect("Database Host Address must be provided through env var!");
    let db = DB::new(&format!("http://{}", db_addr));

    rocket::ignite()
        .manage(db)
        .mount("/", routes![root, post_device])
        .mount("/health", routes![health_check])
        .mount("/dump", routes![dump])
        .mount("/time", routes![time_interval])
        .mount("/sensor", routes![sensor])
        .launch();
}

