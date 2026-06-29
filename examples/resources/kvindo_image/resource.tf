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

# Take a snapshot of the VM disk
resource "kvindo_image" "example" {
  metadata = {
    name = "my-vm-snapshot"
  }
  spec = {
    vm_id = kvindo_vm.main.id
  }
}
