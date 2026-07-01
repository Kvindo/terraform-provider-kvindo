resource "kvindo_vpc_peering" "main" {
  metadata = {
    name = "my-vpc-peering"
  }
}

resource "kvindo_ssh_private_key" "deploy" {
  metadata = {
    name = "deploy-key"
  }
  spec = {
    private_key = file("~/.ssh/id_rsa")
  }
}

resource "kvindo_vpc_peering_external_peer" "example" {
  metadata = {
    name = "external-peer"
  }
  spec = {
    vpc_peering_id     = kvindo_vpc_peering.main.id
    ssh_ipv4          = "203.0.113.10"
    ssh_user           = "ubuntu"
    ssh_private_key_id = kvindo_ssh_private_key.deploy.id
    ipv4_cidrs        = ["192.168.0.0/24"]
  }
}
