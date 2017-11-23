provider "nomad" {
  address = "localhost"
}

# Register a job
resource "nomad_job" "zones" {
  jobspec = "${file("./zones.hcl")}"
}
