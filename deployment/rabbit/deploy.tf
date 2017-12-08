provider "nomad" {
  address = "http://localhost:4646"
}

# Register a job
resource "nomad_job" "rabbit" {
  jobspec = "${file("./rabbit.hcl")}"
}
