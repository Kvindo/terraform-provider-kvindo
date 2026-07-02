resource "kvindo_vpc" "main" {
  metadata = {
    name = "my-vpc"
  }
  spec = {
    hosting_provider_id = "01abc123def456gh789012345"
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

resource "kvindo_kubernetes" "example" {
  metadata = {
    name = "my-cluster"
  }
  spec = {
    version = "1.30"
    control_plane_locations = [
      {
        vpc_subnet_id = kvindo_vpc_subnet.main.id
      }
    ]
    assign_public_ipv4 = true
  }
}

output "api_server_url" {
  value = kvindo_kubernetes.example.status.api_server_url
}
