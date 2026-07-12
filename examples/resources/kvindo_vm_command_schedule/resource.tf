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

# A schedule is a standalone resource; attach it to one or more VMs via the
# VM's command_schedule_ids, not the other way around.
resource "kvindo_vm_command_schedule" "nightly_cleanup" {
  metadata = {
    name = "nightly-cleanup"
  }
  spec = {
    enabled                 = true
    schedule_format         = "cron"
    schedule                = "0 4 * * *"
    command                 = "journalctl --vacuum-time=7d && apt-get clean"
    command_timeout_seconds = 300
  }
}

resource "kvindo_vm" "main" {
  metadata = {
    name = "my-vm"
  }
  spec = {
    offer_id             = "01vm0ffr123456789012345"
    image_id             = "01img123456789012345678"
    vpc_subnet_id        = kvindo_vpc_subnet.main.id
    command_schedule_ids = [kvindo_vm_command_schedule.nightly_cleanup.id]
  }
}
