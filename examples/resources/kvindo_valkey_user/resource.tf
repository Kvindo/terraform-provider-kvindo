resource "kvindo_valkey" "main" {
  metadata = {
    name = "main"
  }
  spec = {
    replicas_per_shard = 1
    valkey_version     = "8.1.9"
    vm_offer_id        = "g3-1c1-100"
    volume_offer_id    = "gp3-750"
    volume_size_gib    = 10
    shards = [
      { name = "shard-0", vpc_subnet_id = data.kvindo_vpc_subnet.app.metadata.id },
    ]
  }
}

# ACL user scoped to a single key prefix, no admin/all-keys access
resource "kvindo_valkey_user" "app" {
  metadata = {
    name = "app-user"
  }
  spec = {
    valkey_id    = kvindo_valkey.main.id
    password     = "change-me-please"
    enabled      = true
    key_patterns = ["cache:*"]
    categories   = ["read", "write"]
    channels     = ["notify:*"]
  }
}

data "kvindo_vpc_subnet" "app" {
  name = "app-subnet"
}
