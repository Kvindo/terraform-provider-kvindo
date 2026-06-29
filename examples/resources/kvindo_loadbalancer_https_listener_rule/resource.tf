resource "kvindo_certificate" "main" {
  metadata = { name = "my-cert" }
  spec = {
    certificate_pem = file("cert.pem")
    private_key_pem = file("key.pem")
  }
}

resource "kvindo_loadbalancer" "main" {
  metadata = { name = "my-lb" }
}

resource "kvindo_loadbalancer_target_group" "main" {
  metadata = { name = "my-tg" }
  spec     = { loadbalancer_id = kvindo_loadbalancer.main.id }
}

resource "kvindo_loadbalancer_https_listener" "main" {
  metadata = { name = "https-443" }
  spec = {
    loadbalancer_id = kvindo_loadbalancer.main.id
    ports           = ["443"]
    tls             = { certificate_id = kvindo_certificate.main.id }
  }
}

resource "kvindo_loadbalancer_https_listener_rule" "example" {
  metadata = { name = "api-route" }
  spec = {
    https_listener_id = kvindo_loadbalancer_https_listener.main.id
    order             = 1

    match = {
      path            = "/api/"
      path_match_type = "prefix"
    }

    forward_to_http_response_action = {
      target_group_id = kvindo_loadbalancer_target_group.main.id
    }
  }
}
