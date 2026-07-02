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

resource "kvindo_vpc_peering" "main" {
  metadata = {
    name = "my-vpc-peering"
  }
}

resource "kvindo_vpc_peering_peer" "example" {
  metadata = {
    name = "peer-a"
  }
  spec = {
    vpc_peering_id = kvindo_vpc_peering.main.id
    vpc_subnet_id  = kvindo_vpc_subnet.main.id
  }
}
