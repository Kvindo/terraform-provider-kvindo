resource "kvindo_security_group" "example" {
  metadata = {
    name = "my-sg"
  }
  spec = {
    # optional, defaults to "deny" (unmatched traffic dropped); set to "allow" for
    # blacklist mode (unmatched traffic allowed, only explicit deny rules block it)
    default_action = "deny"
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
