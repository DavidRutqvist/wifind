job "zones" {
  datacenters = ["dc1"]
  type        = "service"

  update {
    max_parallel = 1
  }

  group "zones" {
    count = 1
    restart {
      attempts = 10
      interval = "5m"
      delay    = "25s"
      mode     = "delay"
    }

    task "mongo" {
      driver = "docker"

      config {
        image = "mongo:3.0"
        volumes = [
          "/mnt/zones/mongo:/data/db"
        ]

        port_map {
          mongo = 27017
        }
      }

      resources {
        cpu    = 256 # 256 MHz
        memory = 1024 # 1GB

        network {
          mbits = 10

          port "mongo" {}
        }
      }
    }

    task "service" {
      driver = "docker"

      config {
        image = "docker.adventic.se/wifind/zones:1.0.2"

        port_map {
          http = 8080
        }

        ssl = true
        auth {
              username = "nomad"
              password = "nomad"
              server_address = "docker.adventic.se"
        }
      }

      env {
        "MONGO_ADDRESS" = "${NOMAD_ADDR_mongo_mongo}"
        "CONSUL_ADDRESS" = "${attr.unique.network.ip-address}:8500"
        "EXCHANGE_TOPIC" = "event"
      }

      resources {
        cpu    = 128 # 128 MHz
        memory = 128 # 128MB

        network {
          mbits = 10

          port "http" {}
        }
      }

      service {
        name = "zones"
        tags = ["http"]
        port = "http"

        check {
          name     = "HTTP Health Check"
          type     = "http"
          interval = "10s"
          timeout  = "2s"
          port     = "http"
          path     = "/"
        }
      }
    }
  }
}
