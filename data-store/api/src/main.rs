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

#[get("/")]
fn dump(db: State<DB>) -> Json<serde_json::Value> {

    Json(db.dump())
}

#[get("/")]
fn root() -> &'static str {
    "This is an API"
}

fn main() {

    let db = DB::new("http://localhost:8086");

    rocket::ignite()
        .manage(db)
        .mount("/", routes![root, post_device])
        .mount("/dump", routes![dump])
        .mount("/time", routes![time_interval])
        //.mount("/device", routes![get_device, post_device])
        .launch();
}

