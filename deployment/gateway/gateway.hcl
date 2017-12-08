job "gateway" {
  datacenters = ["dc1"]
  type        = "service"

  update {
    max_parallel = 1
  }

  group "gateway" {
    count = 2

    restart {
      attempts = 10
      interval = "5m"
      delay    = "25s"
      mode     = "delay"
    }

    task "gateway" {
      driver = "docker"

      config {
        image = "docker.adventic.se/wifind/gateway:0.0.4"

        port_map {
          http = 3000
        }
        ssl = true
        auth {
              username = "nomad"
              password = "nomad"
              server_address = "docker.adventic.se"
        }
      }

      env {
        "CONSUL_ADDR" = "${attr.unique.network.ip-address}:8500"
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
        name = "gateway"
        tags = ["http", "urlprefix-api.wifind.se:9999/"]
        port = "http"

        check {
          name     = "HTTP Health Check"
          type     = "http"
          interval = "10s"
          timeout  = "2s"
          port     = "http"
          path     = "/api"
        }
      }
    }
  }
}
