
use influx_db_client::{Client, Point, Value, Series};
use serde_json;

use std::collections::HashMap;

static _DB: &'static str = "db1";

pub struct DB {
    cl: Client
}

impl DB {

    pub fn new(host: &str) -> DB {
        let mut client = Client::new(host, _DB);

        client.create_database(_DB);
        client.swith_database(_DB);

        DB { cl: client }
    }

    pub fn insert(&self,
                  device_hash: String,
                  sensor: String,
                  rssi: i32,
                  timestamp: u64) {

        let mut point = point!("pkt_txn");
        point.add_timestamp(timestamp as i64);
        point.add_tag("sensor", Value::String(sensor));
        point.add_field("device_hash", Value::String(device_hash));
        point.add_field("rssi", Value::Integer(rssi as i64));

        self.cl.write_point(point, None, None);
    }

    pub fn is_responding(&self) -> bool {
        self.cl.ping()
    }

    fn flatten_series(series: &Series) -> Vec<HashMap<String, serde_json::Value>> {

        let mut result = Vec::new();
        for measurement in series.values.iter() {
            let key_vals = series.columns
                .iter()
                .map(|s| s.to_string())
                .zip(measurement
                     .iter()
                     .map(|val| json!(val))
                 ).collect::<HashMap<String, serde_json::Value>>();

            result.push(key_vals);
        }

        result
    }

    fn query(&self, q: &str) -> Vec<HashMap<String, serde_json::Value>> {

        let found = match self.cl.query(q, None) {
            Ok(Some(found)) => found,
            _ => Vec::new()
        };

        let flattened = found
            .into_iter()
            .flat_map(|find| find.series.unwrap())
            .collect::<Vec<_>>();

        let result = flattened
            .iter()
            .flat_map(|series| Self::flatten_series(series))
            .collect::<Vec<HashMap<String, serde_json::Value>>>();

        result
    }

    pub fn dump(&self) -> serde_json::Value {
        let result = self.query("select * from pkt_txn");

        json!(result)
    }

    pub fn get_time_interval(&self, start: u64, stop: u64) -> serde_json::Value {
        let result = self.query(
            &format!("select * from pkt_txn where ({} < time and time < {})",
                start,
                stop)
        );

        json!(result)
    }

    pub fn get_sensor(&self, sensor_id: String) -> serde_json::Value {
        let result = self.query(
            &format!("select * from pkt_txn where sensor='{}'",
                sensor_id)
        );

        json!(result)
    }
}

