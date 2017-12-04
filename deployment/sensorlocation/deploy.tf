provider "nomad" {
  address = "http://localhost:4646"
}

# Register a job
resource "nomad_job" "sensorlocation" {
  jobspec = "${file("./sensorlocation.hcl")}"
}
