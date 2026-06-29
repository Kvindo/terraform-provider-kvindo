resource "kvindo_open_vpn" "main" {
  metadata = {
    name = "my-vpn"
  }
}

resource "kvindo_open_vpn_user" "example" {
  metadata = {
    name = "alice"
  }
  spec = {
    open_vpn_id = kvindo_open_vpn.main.id
  }
}

output "vpn_config" {
  value     = kvindo_open_vpn_user.example.status.config
  sensitive = true
}
