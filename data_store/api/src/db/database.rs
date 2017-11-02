
use bson::Document;
use mongodb::{Client, ThreadedClient};
use mongodb::db::ThreadedDatabase;
use mongodb::coll::Collection;

pub struct DB {
    coll: Collection
}

impl DB {
    pub fn new(host: &str, port: u16, _db: &str, _coll: &str) -> DB {
        let client = Client::connect(host, port)
            .expect("Failed to initialize DB");

        let coll = client.db(_db).collection(_coll);
        DB { coll: coll }
    }

    pub fn insert(&self, doc: Document) {
        self.coll.insert_one(doc, None)
            .expect("Failed to make insertion!");
    }

    pub fn get(&self, filter: Document, result: &mut Vec<Document>) {
        let cursor = self.coll.find(Some(filter), None)
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

