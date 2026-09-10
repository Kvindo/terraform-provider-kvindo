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

resource "kvindo_postgresql_user" "app" {
  metadata = {
    name = "app-user"
  }
  spec = {
    postgre_sql_id   = kvindo_postgresql.main.id
    login            = true
    connection_limit = 20
    # Left unset here: the platform generates a random password on create, which never appears
    # in state or plan output. Set explicitly to pin a known password instead.
  }
}

data "kvindo_vpc_subnet" "app" {
  name = "app-subnet"
}
