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
# VM's on_off_maintenance_action_ids, not the other way around.
resource "kvindo_vm_on_off_maintenance_action" "evening_stop" {
  metadata = {
    name = "evening-stop"
  }
  spec = {
    enabled         = true
    schedule_format = "cron"
    schedule        = "0 20 * * *"
    target_state    = "stopped"
  }
}

resource "kvindo_vm" "main" {
  metadata = {
    name = "my-vm"
  }
  spec = {
    offer_id                     = "01vm0ffr123456789012345"
    image_id                     = "01img123456789012345678"
    vpc_subnet_id                = kvindo_vpc_subnet.main.id
    on_off_maintenance_action_ids = [kvindo_vm_on_off_maintenance_action.evening_stop.id]
  }
}
