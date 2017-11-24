
use std::thread;
use bson::Document;
use mongodb::{Client, ThreadedClient};
use mongodb::db::ThreadedDatabase;

static _DB: &'static str = "db1";
static _COLL: &'static str = "devices";

pub struct DB {
    cl: Client
}

impl DB {
    pub fn new(host: &str, port: u16) -> DB {
        let client = Client::connect(host, port)
            .expect("Failed to initialize DB");

        DB { cl: client }
    }

    pub fn insert(&self, doc: Document) {
        let coll = self.cl.clone().db(_DB).collection(_COLL);

        thread::spawn(move || {
            coll.insert_one(doc, None)
                .expect("Failed to make insertion!");
        });
    }

    pub fn dump(&self, result: &mut Vec<Document>) {
        let coll = self.cl.db(_DB).collection(_COLL);
        let cursor = coll.find(None, None)
            .ok().expect("Failed to execute find.");

        for item in cursor {

            match item {
                Ok(mut doc) => {
                    doc.remove("_id");
                    result.push(doc);
                },
                Err(_) => panic!("Failed to get result from server!"),
            }

        }
    }

    pub fn get(&self, filter: Document, result: &mut Vec<Document>) {
        let coll = self.cl.db(_DB).collection(_COLL);
        let cursor = coll.find(Some(filter), None)
            .ok().expect("Failed to execute find.");

        for item in cursor {

            match item {
                Ok(mut doc) => {
                    doc.remove("_id");
                    result.push(doc);
                },
                Err(_) => panic!("Failed to get result from server!"),
            }

        }
    }

}

