resource "kvindo_postgresql" "main" {
  metadata = {
    name = "main"
  }
  spec = {
    version                  = "16"
    replicas_per_shard_group = 1
    vm_offer_id              = "g3-1c1-100"
    volume_offer_id          = "gp3-750"
    volume_size_gib          = 20
    shard_groups = [
      { is_coordinator = true, name = "shard-0", vpc_subnet_id = data.kvindo_vpc_subnet.app.metadata.id },
    ]
  }
}

data "kvindo_vpc_subnet" "app" {
  name = "app-subnet"
}
