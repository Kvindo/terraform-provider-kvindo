data "kvindo_folder" "default" {
  name = "default"
}

data "kvindo_vpc_subnet" "app" {
  name = "app-subnet"
}

data "kvindo_clickhouse_parameters_set" "analytics" {
  name = "analytics"
}

resource "kvindo_clickhouse" "analytics" {
  metadata = {
    name      = "analytics-cluster"
    folder_id = data.kvindo_folder.default.metadata.id
  }
  spec = {
    clickhouse_version       = "26.8.11.7"
    replicas_per_shard_group = 1
    parameters_set_id         = data.kvindo_clickhouse_parameters_set.analytics.metadata.id
    vm_offer_id               = "g3-1c1-100"
    volume_offer_id           = "gp3-750"
    volume_size_gib           = 20
    keeper_vm_offer_id        = "g3-1c1-100"
    keeper_volume_offer_id    = "gp3-750"
    keeper_volume_size_gib    = 10
    backup_retention_days     = 7
    full_backup_interval_days = 7
    shard_groups = [
      { name = "shard-0", vpc_subnet_id = data.kvindo_vpc_subnet.app.metadata.id }
    ]
    keepers = [
      { name = "keeper-0", vpc_subnet_id = data.kvindo_vpc_subnet.app.metadata.id }
    ]
  }
}
