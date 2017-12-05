job "rabbit" {
  datacenters = ["dc1"]
  type        = "service"

  update {
    max_parallel = 1
  }

  group "rabbit" {
    constraint {
      distinct_hosts = true
    }

    restart {
      attempts = 10
      interval = "5m"
      delay    = "25s"
      mode     = "delay"
    }

    task "rabbit" {
      driver = "docker"

      config {
        image = "rabbitmq:3-management-alpine"

        volumes = [
          "/mnt/rabbit:/var/lib/rabbitmq/mnesia"
        ]

        port_map {
          amqp = 5672
          management = 15672
        }
      }

      resources {
        cpu    = 500 # 500 MHz
        memory = 256 # 256MB

        network {
          mbits = 10

          port "amqp" {
            static = 5672
          }

          port "management" {
            static = 15672
          }
        }
      }

      service {
        name = "rabbit"
        tags = ["amqp"]
        port = "amqp"

        check {
          name     = "AMQP TCP Connection"
          type     = "tcp"
          interval = "10s"
          timeout  = "2s"
          port      = "amqp"
        }

        check {
          name     = "Management GUI"
          type     = "http"
          interval = "10s"
          timeout  = "2s"
          port     = "management"
          path     = "/"
        }
      }
    }
  }
}