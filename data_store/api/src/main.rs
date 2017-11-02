#![feature(plugin, decl_macro)]
#![plugin(rocket_codegen)]

#[macro_use]
extern crate serde_derive;
extern crate serde;
#[macro_use(bson, doc)]
extern crate bson;
extern crate mongodb;
extern crate rocket;
extern crate rocket_contrib;

use bson::Document;

use rocket::State;
use rocket_contrib::Json;

mod db;
use db::database::DB;

#[derive(Deserialize)]
struct Data {
    device: String,
    sensor: String,
    rssi: i32,
    timestamp: String, // TODO: Should be changed to TimeDate type
}

#[get("/<device>")]
fn device(device: String, db: State<DB>) -> Json<Vec<Document>> {
    let filter = doc! {"device": &device};

    let mut results = Vec::new();
    db.get(filter, &mut results);

    Json(results)
}

#[post("/insert", data = "<data>")]
fn insert(data: Json<Data>, db: State<DB>) -> &'static str {
    // let data = match bson::to_bson(&data) { Ok(serialized) => serialized, Err(e) => return format!("Serialization error: {}", e), }};

    let doc = doc! {
        "device": data.device.clone(),
        "sensor": data.sensor.clone(),
        "rssi": data.rssi.clone(),
        "timestamp": data.timestamp.clone(),
    };

    db.insert(doc);

    return "success"

}

#[get("/")]
fn all(db: State<DB>) -> Json<Vec<Document>> {
    let filter = doc! {
        "device": "*"
    };

    let mut results = Vec::new();
    db.get(filter, &mut results);

    Json(results)
}

fn main() {

    let db = DB::new("localhost", 27017, "db1", "devices");

    rocket::ignite()
        .manage(db)
        .mount("/", routes![all, insert])
        .mount("/device", routes![device])
        .launch();
}
