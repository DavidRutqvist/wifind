
extern crate consul;

use self::consul::Client;
use std::collections::HashMap;
use std::env;

pub fn get_services(name: &str) -> Option<Vec<String>> {
    let server = env::var("CONSUL_ADDR")
        .unwrap_or("130.240.5.73:8500".to_string());

    let client = Client::new(&format!("http://{}", server));
    let services: HashMap<String, Vec<String>> = client.catalog.services().unwrap();

    match services.get(name) {
        Some(vec) => Some(vec.iter().map(|v| v.clone()).collect::<Vec<_>>()),
        None => None
    }
}

