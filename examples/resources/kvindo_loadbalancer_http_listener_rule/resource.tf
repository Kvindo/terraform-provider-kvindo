resource "kvindo_loadbalancer" "main" {
  metadata = { name = "my-lb" }
}

resource "kvindo_loadbalancer_target_group" "main" {
  metadata = { name = "my-tg" }
  spec     = { loadbalancer_id = kvindo_loadbalancer.main.id }
}

resource "kvindo_loadbalancer_http_listener" "main" {
  metadata = { name = "http-80" }
  spec = {
    loadbalancer_id = kvindo_loadbalancer.main.id
    ports           = ["80"]
  }
}

# A rule matches requests, then applies exactly one action block.
resource "kvindo_loadbalancer_http_listener_rule" "example" {
  metadata = { name = "api-route" }
  spec = {
    http_listener_id = kvindo_loadbalancer_http_listener.main.id
    order            = 1

    match = {
      path            = "/api/"
      path_match_type = "prefix"
    }

    forward_to_http_response_action = {
      target_group_id = kvindo_loadbalancer_target_group.main.id
    }
  }
}

# Other action blocks (pick one per rule):
#   static_response_action       = { status_code = 200, content_type = "text/plain", body_string = "ok" }
#   forward_to_https_response_action = { target_group_id = "...", to_ports = ["8443"], tls = { verify = true, sni_server_name = "backend.internal" } }
#   path_rewrite_action          = { source_path = "/old", destination_path = "/new", path_type = "prefix" }
#   set_request_headers_action   = { headers = { "X-Env" = "prod" } }
#   delete_response_headers_action = { headers = ["Server"] }
