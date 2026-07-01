resource "kvindo_loadbalancer" "main" {
  metadata = { name = "my-lb" }
}

resource "kvindo_loadbalancer_http_listener" "example" {
  metadata = { name = "http-80" }
  spec = {
    loadbalancer_id = kvindo_loadbalancer.main.id
    ports           = ["80"]

    # Optional ordered allow/deny rules for inbound traffic.
    security_rules = [
      {
        order        = 1
        action       = "allow"
        description  = "office network"
        ipv4_blocks = ["203.0.113.0/24"]
      }
    ]
  }
}
