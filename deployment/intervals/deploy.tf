provider "nomad" {
  address = "http://localhost:4646"
}

# Register a job
resource "nomad_job" "intervals" {
  jobspec = "${file("./intervals.hcl")}"
}