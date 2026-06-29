resource "kvindo_loadbalancer" "main" {
  metadata = { name = "my-lb" }
}

resource "kvindo_loadbalancer_tcp_listener" "example" {
  metadata = { name = "tcp-5432" }
  spec = {
    loadbalancer_id = kvindo_loadbalancer.main.id
    ports           = ["5432"]

    security_rules = [
      {
        order        = 1
        action       = "allow"
        ip_v4_blocks = ["10.0.0.0/8"]
      }
    ]
  }
}
