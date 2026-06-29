resource "kvindo_loadbalancer" "main" {
  metadata = { name = "my-lb" }
}

resource "kvindo_loadbalancer_target_group" "main" {
  metadata = { name = "my-tg" }
  spec     = { loadbalancer_id = kvindo_loadbalancer.main.id }
}

resource "kvindo_loadbalancer_udp_listener" "main" {
  metadata = { name = "udp-53" }
  spec = {
    loadbalancer_id = kvindo_loadbalancer.main.id
    ports           = ["53"]
  }
}

resource "kvindo_loadbalancer_udp_listener_rule" "example" {
  metadata = { name = "dns-forward" }
  spec = {
    udp_listener_id = kvindo_loadbalancer_udp_listener.main.id
    order           = 1

    forward_to_udp_response_action = {
      target_group_id = kvindo_loadbalancer_target_group.main.id
    }
  }
}
