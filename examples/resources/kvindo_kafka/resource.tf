data "kvindo_folder" "default" {
  name = "default"
}

data "kvindo_vpc_subnet" "brokers" {
  name = "brokers-subnet"
}

data "kvindo_vpc_subnet" "controllers" {
  name = "controllers-subnet"
}

resource "kvindo_kafka" "main" {
  metadata = {
    name      = "main"
    folder_id = data.kvindo_folder.default.metadata.id
  }
  spec = {
    kafka_version = "4.3.1"

    # One sizing triplet per ROLE: brokers hold tenant data, controllers hold only the
    # metadata quorum. A role's nodes are interchangeable by design, so sizing is per role.
    broker_vm_offer_id         = "g3-2c4-100"
    broker_volume_offer_id     = "gp3-750"
    broker_volume_size_gib     = 20
    controller_vm_offer_id     = "g3-1c1-100"
    controller_volume_offer_id = "gp3-750"
    controller_volume_size_gib = 10

    # Controllers must be exactly 1 or 3 - the KRaft quorum needs an odd voter count.
    brokers = [
      { name = "broker-0", vpc_subnet_id = data.kvindo_vpc_subnet.brokers.metadata.id },
    ]
    controllers = [
      { name = "controller-0", vpc_subnet_id = data.kvindo_vpc_subnet.controllers.metadata.id },
    ]
  }
}
