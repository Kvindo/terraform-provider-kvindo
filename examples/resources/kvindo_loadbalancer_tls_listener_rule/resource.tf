resource "kvindo_certificate" "backend_ca" {
  metadata = { name = "backend-ca" }
  spec = {
    certificate_pem = file("ca.pem")
    private_key_pem = file("ca-key.pem")
  }
}

resource "kvindo_loadbalancer" "main" {
  metadata = { name = "my-lb" }
}

resource "kvindo_loadbalancer_target_group" "main" {
  metadata = { name = "my-tg" }
  spec     = { loadbalancer_id = kvindo_loadbalancer.main.id }
}

resource "kvindo_loadbalancer_tls_listener" "main" {
  metadata = { name = "tls-443" }
  spec = {
    loadbalancer_id = kvindo_loadbalancer.main.id
    ports           = ["443"]
    tls             = { certificate_id = kvindo_certificate.backend_ca.id }
  }
}

resource "kvindo_loadbalancer_tls_listener_rule" "example" {
  metadata = { name = "tls-forward" }
  spec = {
    tls_listener_id = kvindo_loadbalancer_tls_listener.main.id
    order           = 1

    # Re-encrypt to a backend that presents its own certificate.
    forward_to_tls_response_action = {
      target_group_id = kvindo_loadbalancer_target_group.main.id
      tls = {
        verify            = true
        sni_server_name   = "backend.internal"
        ca_certificate_id = kvindo_certificate.backend_ca.id
      }
    }
  }
}
