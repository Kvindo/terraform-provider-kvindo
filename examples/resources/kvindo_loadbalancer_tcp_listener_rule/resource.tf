resource "kvindo_loadbalancer" "main" {
  metadata = { name = "my-lb" }
}

resource "kvindo_loadbalancer_target_group" "main" {
  metadata = { name = "my-tg" }
  spec     = { loadbalancer_id = kvindo_loadbalancer.main.id }
}

resource "kvindo_loadbalancer_tcp_listener" "main" {
  metadata = { name = "tcp-5432" }
  spec = {
    loadbalancer_id = kvindo_loadbalancer.main.id
    ports           = ["5432"]
  }
}

# TCP rules forward to a target group (optionally re-mapping ports).
resource "kvindo_loadbalancer_tcp_listener_rule" "example" {
  metadata = { name = "db-forward" }
  spec = {
    tcp_listener_id = kvindo_loadbalancer_tcp_listener.main.id
    order           = 1

    forward_to_tcp_response_action = {
      target_group_id = kvindo_loadbalancer_target_group.main.id
    }
  }
}

# To terminate/originate TLS to the backend instead, use:
#   forward_to_tls_response_action = {
#     target_group_id = "..."
#     tls = { verify = true, sni_server_name = "backend.internal", ca_certificate_id = "..." }
#   }
