resource "kvindo_quota" "example" {
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
