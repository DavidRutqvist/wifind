provider "nomad" {
  address = "http://130.240.5.73:4646"
}

# Register a job
resource "nomad_job" "datastore" {
  jobspec = "${file("./datastore.hcl")}"
}
