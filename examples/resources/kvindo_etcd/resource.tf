resource "kvindo_etcd" "main" {
  metadata = {
    name = "main"
  }
  spec = {
    etcd_version    = "3.5.15"
    tier            = "standard"
    vm_offer_id     = "g3-1c1-100"
    volume_offer_id = "gp3-750"
    volume_size_gib = 10
    instances = [
      { id = "01abc123def456gh789012345", vpc_subnet_id = data.kvindo_vpc_subnet.app.metadata.id },
    ]
  }
}

data "kvindo_vpc_subnet" "app" {
  name = "app-subnet"
}
