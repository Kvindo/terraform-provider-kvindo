# kvindo_loadbalancer_target_group has no spec of its own - see its own example for why.
resource "kvindo_loadbalancer_target_group" "main" {
  metadata = {
    name = "my-tg"
  }
}

resource "kvindo_loadbalancer_target_group_service_discovery_target" "example" {
  metadata = {
    name = "k8s-discovery"
  }
  spec = {
    target_group_id = kvindo_loadbalancer_target_group.main.id
    label_selectors = {
      app = "backend"
      env = "production"
    }
  }
}
