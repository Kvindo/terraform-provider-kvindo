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

resource "kvindo_kubernetes" "main" {
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
  }
}

resource "kvindo_kubernetes_node_group" "example" {
  metadata = {
    name = "workers"
  }
  spec = {
    kubernetes_id      = kvindo_kubernetes.main.id
    vpc_subnet_id      = kvindo_vpc_subnet.main.id
    vm_offer_id        = "01vm0ffr123456789012345"
    desired_node_count = 3
  }
}
