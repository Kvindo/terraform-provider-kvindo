resource "kvindo_security_group" "example" {
  metadata = {
    name = "my-sg"
  }
  spec = {
    ingress = [
      {
        ports       = ["tcp:22", "tcp:80", "tcp:443", "icmp"]
        ipv4_blocks = ["external"]
        action      = "allow"
      },
      {
        ports       = ["all"]
        ipv4_blocks = ["local"]
        action      = "allow"
      }
    ]
    egress = [
      {
        ports       = ["all"]
        ipv4_blocks = ["all"]
        action      = "allow"
      }
    ]
  }
}
