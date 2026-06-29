resource "kvindo_quota" "vm_cpu" {
  metadata = {
    name = "vm-cpu-quota"
  }
  spec = {
    product   = "compute"
    resource  = "vm"
    parameter = "cpu"
    limit     = 100
  }
}

# Request an increase of the quota limit
resource "kvindo_quota_change_request" "example" {
  metadata = {
    name = "increase-vm-cpu-to-200"
  }
  spec = {
    quota_id        = kvindo_quota.vm_cpu.id
    new_quota_limit = 200
  }
}
