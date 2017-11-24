
extern crate consul;
extern crate pretty_env_logger;

use self::consul::Client;
use std::env;

pub fn get_node_address(name: &str) -> Option<String> {

    let server = env::var("CONSUL_ADDR")
        .unwrap_or("130.240.5.73:8500".to_string());

    println!("Fetching {} from Consul at http://{}", name, server);
    let client = Client::new(&format!("http://{}", server));
    let nodes = client.catalog.get_nodes(name.to_string())
        .expect("Fetching of nodes failed!");

    match nodes.first() {
        Some(node) => Some(format!("{}:{}", node.Address, node.ServicePort)),
        None => None
    }
}

