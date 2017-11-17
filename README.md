# ![alt text](logo.png "WiFind Logo") WiFind

The WiFind project is part of the course in Pervasive Computing (M7012E) at Luleå University of Technology.

## Links
- [Webpage](http://wifind.se)
- [Application](http://app.wifind.se:9999)
- [Nomad UI](http://app.wifind.se:4646/ui)
- [Consul UI](http://app.wifind.se:8500/ui)

## Modules
Below is a description of the modules stored in this repository

### Frontend
The frontend application is an Angular application for displaying a user interface to the end user. It is the single point of contact for end users and is used to setup the system, monitor zones and provide analysis insights. See readme in separate folder for instructions for the tool chain.

### Data Store

### Deployment
The deployment folder contains separate folders for each module that is deployed in the complete system. Each folder contains a Terraform job description and in most cases a Nomad template. To deploy a module CD to the folder and run `terraform init` followed by `terraform plan` and then `terraform apply`.

### Development Environment
The development environment folder contains a Vagrantfile for deploying a replica of the production environment locally on the developer's computer. The configuration supports various providers which covers all major operating systems. Just CD to the folder and run `vagrant up` to set up the environment.