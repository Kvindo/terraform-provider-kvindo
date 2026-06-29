resource "kvindo_security_group" "example" {
  metadata = {
    name = "my-sg"
  }
  spec = {
    ingress = [
      {
        ports       = ["22", "80", "443"]
        ipv4_blocks = ["0.0.0.0/0"]
        action      = "allow"
      }
    ]
    egress = [
      {
        ports       = []
        ipv4_blocks = ["0.0.0.0/0"]
        action      = "allow"
      }
    ]
  }
}
