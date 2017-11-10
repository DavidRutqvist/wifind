provider "nomad" {
  address = "http://localhost:4646"
}

# Register a job
resource "nomad_job" "app" {
  jobspec = "${file("./app.hcl")}"
}
