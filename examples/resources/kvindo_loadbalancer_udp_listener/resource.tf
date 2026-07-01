resource "kvindo_loadbalancer" "main" {
  metadata = { name = "my-lb" }
}

resource "kvindo_loadbalancer_udp_listener" "example" {
  metadata = { name = "udp-53" }
  spec = {
    loadbalancer_id = kvindo_loadbalancer.main.id
    ports           = ["53"]

    security_rules = [
      {
        order        = 1
        action       = "allow"
        ipv4_blocks = ["0.0.0.0/0"]
      }
    ]
  }
}
