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

resource "kvindo_loadbalancer_tls_listener" "example" {
  metadata = { name = "tls-443" }
  spec = {
    loadbalancer_id = kvindo_loadbalancer.main.id
    ports           = ["443"]

    tls = {
      certificate_id = kvindo_certificate.main.id
      protocols      = ["TLSv1.2", "TLSv1.3"]
    }
  }
}
