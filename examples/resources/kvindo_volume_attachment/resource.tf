resource "kvindo_vpc" "main" {
  metadata = {
    name = "my-vpc"
  }
  spec = {
    hosting_provider_id = "01abc123def456gh789012345"
    ipv4_cidr           = "10.0.0.0/16"
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

resource "kvindo_vm" "main" {
  metadata = {
    name = "my-vm"
  }
  spec = {
    offer_id      = "01vm0ffr123456789012345"
    image_id      = "01img123456789012345678"
    vpc_subnet_id = kvindo_vpc_subnet.main.id
  }
}

resource "kvindo_volume" "main" {
  metadata = {
    name = "my-volume"
  }
  spec = {
    hosting_provider_id = "01abc123def456gh789012345"
    offer_id            = "01vol0ffr12345678901234"
    size_gib            = 50
  }
}

resource "kvindo_volume_attachment" "example" {
  metadata = {
    name = "my-attachment"
  }
  spec = {
    vm_id     = kvindo_vm.main.id
    volume_id = kvindo_volume.main.id
  }
}
