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

# Take a snapshot of the VM disk
resource "kvindo_image" "example" {
  metadata = {
    name = "my-vm-snapshot"
  }
  spec = {
    vm_id = kvindo_vm.main.id
  }
}

# Or back up a standalone volume instead - exactly one of vm_id/volume_id must be
# set. Only a volume-sourced image (status.is_vm_image = false) can later be used
# to restore a kvindo_volume via its image_id field.
resource "kvindo_volume" "data" {
  metadata = {
    name = "my-data-volume"
  }
  spec = {
    hosting_provider_id = "01abc123def456gh789012345"
    offer_id            = "01vol0ffr12345678901234"
    size_gib            = 50
  }
}

resource "kvindo_image" "volume_backup" {
  metadata = {
    name = "my-volume-backup"
  }
  spec = {
    volume_id = kvindo_volume.data.id
  }
}
