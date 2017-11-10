job "fabio" {
  datacenters = ["dc1"]
  type = "system"
  update {
    stagger = "5s"
    max_parallel = 1
  }

  group "fabio" {
    task "fabio" {
      driver = "exec"
      config {
        command = "fabio-1.5.3-go1.9.2-linux_amd64"
      }

      artifact {
        source = "https://github.com/fabiolb/fabio/releases/download/v1.5.3/fabio-1.5.3-go1.9.2-linux_amd64"
        options {
          checksum = "sha256:ad352a3e770215219c57257c5dcbb14aee83aa50db32ba34431372b570aa58e5"
        }
      }

      resources {
        cpu = 500
        memory = 64
        network {
          mbits = 1

          port "http" {
            static = 9999
          }
          port "ui" {
            static = 9998
          }
        }
      }
    }
  }
}