resource "kvindo_vpc" "main" {
  metadata = {
    name = "my-vpc"
  }
  spec = {
    hosting_provider_id = "01abc123def456gh789012345"
    ipv4_cidr           = "10.0.0.0/16"
  }
}

resource "kvindo_vpc_subnet" "main" {
  metadata = {
    name = "my-subnet"
  }
  spec = {
    vpc_id    = kvindo_vpc.main.id
    ipv4_cidr = "10.0.1.0/24"
  }
}

resource "kvindo_loadbalancer" "example" {
  metadata = {
    name = "my-lb"
  }
  spec = {
    vpc_subnet_id = kvindo_vpc_subnet.main.id
  }
}

output "public_ip" {
  value = kvindo_loadbalancer.example.status.public_ip_v4
}
