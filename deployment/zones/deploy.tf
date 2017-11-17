provider "nomad" {
  address = "http://srv.wifind.se:4646"
}

# Register a job
resource "nomad_job" "zones" {
  jobspec = "${file("./zones.hcl")}"
}
