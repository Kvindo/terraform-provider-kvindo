# kvindo_loadbalancer_target_group has no spec of its own - see its own example for why.
resource "kvindo_loadbalancer_target_group" "main" {
  metadata = {
    name = "my-tg"
  }
}

resource "kvindo_loadbalancer_target_group_static_target" "example" {
  metadata = {
    name = "backend-1"
  }
  spec = {
    target_group_id = kvindo_loadbalancer_target_group.main.id
    ip_or_hostname  = "10.0.1.10"
  }
}
