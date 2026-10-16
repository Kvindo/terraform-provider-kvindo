data "kvindo_vpc" "main" {
  name = "main"
}

data "kvindo_vpc_subnet" "private_a" {
  name   = "private-a"
  vpc_id = data.kvindo_vpc.main.metadata.id
}

data "kvindo_vpc_subnet" "private_b" {
  name   = "private-b"
  vpc_id = data.kvindo_vpc.main.metadata.id
}

resource "kvindo_opensearch" "search" {
  metadata = {
    name      = "search-cluster"
    folder_id = data.kvindo_vpc.main.metadata.folder_id
  }
  spec = {
    open_search_version = "3.9.0"

    # One sizing triplet per ROLE, never per member: a cluster manager holds the metadata quorum, a
    # data node holds the shards, and they are sized - and rolled - independently.
    data_vm_offer_id        = "g3-2c4-100"
    data_volume_offer_id    = "gp3-750"
    data_volume_size_gib    = 100
    manager_vm_offer_id     = "g3-1c1-100"
    manager_volume_offer_id = "gp3-750"
    manager_volume_size_gib = 10

    # Scale-OUT only: removing a data node is rejected. Draining one needs its shards relocated
    # first, and with no backups a mistake is unrecoverable.
    data_nodes = [
      { name = "data-0", vpc_subnet_id = data.kvindo_vpc_subnet.private_a.metadata.id },
      { name = "data-1", vpc_subnet_id = data.kvindo_vpc_subnet.private_b.metadata.id },
    ]

    # Exactly 1 or 3 - never 2 or 4: a majority is more than half, so an even count tolerates no
    # more failures than the odd count below it and costs more. Membership is append-only and grows
    # 1 -> 3 with BOTH joiners in one request.
    cluster_managers = [
      { name = "manager-0", vpc_subnet_id = data.kvindo_vpc_subnet.private_a.metadata.id },
    ]

    # Immutable after creation, like the version and both volume offers/sizes above.
    create_public_ipv4 = false
    use_fqdn           = false
  }
}
