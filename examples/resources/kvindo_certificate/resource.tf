resource "kvindo_certificate" "example" {
  metadata = {
    name = "my-tls-cert"
  }
  spec = {
    certificate_pem = file("cert.pem")
    private_key_pem = file("key.pem")
  }
}
