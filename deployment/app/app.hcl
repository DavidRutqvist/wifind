job "app" {
  datacenters = ["dc1"]
  type        = "service"

  update {
    max_parallel = 1
  }

  group "app" {
    count = 2

    restart {
      attempts = 10
      interval = "5m"
      delay    = "25s"
      mode     = "delay"
    }

    task "app" {
      driver = "docker"

      config {
        image = "docker.adventic.se/wifind/app:0.0.1"

        port_map {
          http = 80
        }
        ssl = true
        auth {
              username = "nomad"
              password = "nomad"
              server_address = "docker.adventic.se"
        }
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
        name = "app"
        tags = ["http", "urlprefix-app.wifind.se:9999/"]
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