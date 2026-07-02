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

resource "kvindo_ssh_key" "main" {
  metadata = {
    name = "my-key"
  }
  spec = {
    public_key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIHKGbFGS... user@host"
  }
}

resource "kvindo_security_group" "main" {
  metadata = {
    name = "my-sg"
  }
  spec = {
    ingress = [
      {
        ports       = ["22", "80", "443"]
        ipv4_blocks = ["0.0.0.0/0"]
        action      = "allow"
      }
    ]
    egress = [
      {
        ports       = []
        ipv4_blocks = ["0.0.0.0/0"]
        action      = "allow"
      }
    ]
  }
}

resource "kvindo_vm" "example" {
  metadata = {
    name = "my-vm"
  }
  spec = {
    offer_id           = "01vm0ffr123456789012345"
    image_id           = "01img123456789012345678"
    vpc_subnet_id      = kvindo_vpc_subnet.main.id
    ssh_key_ids        = [kvindo_ssh_key.main.id]
    security_group_ids = [kvindo_security_group.main.id]
  }
}

output "private_ip" {
  value = kvindo_vm.example.status.private_ipv4
}

# A VM created from a standalone volume (instead of an image) attaches its boot volume inline via
# spec.boot_volume_attachment. This lets a single apply create the volume, attach it as the boot
# disk, and bring the VM up in state = "running" — kvindo_volume has no dependency on kvindo_vm,
# so there's no ordering issue, and no separate kvindo_volume_attachment resource is needed.
resource "kvindo_volume" "boot" {
  metadata = {
    name = "my-vm-boot-volume"
  }
  spec = {
    hosting_provider_id = "01abc123def456gh789012345"
    offer_id            = "01vol0ffr12345678901234"
    size_gib            = 50
  }
}

resource "kvindo_vm" "from_volume" {
  metadata = {
    name = "my-vm-from-volume"
  }
  spec = {
    offer_id      = "01vm0ffr123456789012345"
    vpc_subnet_id = kvindo_vpc_subnet.main.id
    ssh_key_ids   = [kvindo_ssh_key.main.id]
    vm_state      = "running"
    boot_volume_attachment = {
      volume_id = kvindo_volume.boot.id
    }
  }
}
