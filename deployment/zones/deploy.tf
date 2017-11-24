provider "nomad" {
  address = "localhost:4646"
}

# Register a job
resource "nomad_job" "zones" {
  jobspec = "${file("./zones.hcl")}"
}
