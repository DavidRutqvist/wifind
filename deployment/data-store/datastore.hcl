
job "datastore" {
  datacenters = ["dc1"]
  type        = "service"

  update {
    max_parallel = 1
  }

  group "datastore" {
    count = 1

    restart {
      attempts = 10
      interval = "5m"
      delay    = "25s"
      mode     = "delay"
    }

    task "datastore" {
      driver = "docker"

      config {
        image = "docker.adventic.se/wifind/datastore:1.2"

        ssl = true
        auth {
              username = "nomad"
              password = "nomad"
              server_address = "docker.adventic.se"
        }

        port_map {
            datastore = 80
        }
      }

      resources {
        cpu    = 128 # 128 MHz
        memory = 128 # 128MB

        network {
          mbits = 10

          port "datastore" {}
        }
      }

      env {
        "DB_HOST_ADDR" = "${NOMAD_ADDR_influxdb_influxdb}"
        "ROCKET_ENV" = "production"
        "CONSUL_ADDR" = "${attr.unique.network.ip-address}:8500"
      }

      service {
        name = "datastore"
        tags = ["http"]
        port = "datastore"

        check {
          name     = "HTTP Health Check"
          type     = "http"
          interval = "10s"
          timeout  = "2s"
          port     = "datastore"
          path     = "/health"
        }
      }
    }

    task "influxdb" {
      driver = "docker"

      config {
        image = "influxdb"

        volumes = [
          "/mnt/datastore/influxdb:/var/lib/influxdb"
        ]

        port_map {
            influxdb = 8086
        }
      }

      resources {
        cpu    = 128 # 128 MHz
        memory = 128 # 128MB

        network {
          mbits = 10

          port "influxdb" {}
        }
      }

    }
  }
}
