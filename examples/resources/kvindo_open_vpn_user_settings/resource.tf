resource "kvindo_open_vpn" "main" {
  metadata = {
    name = "my-vpn"
  }
}

resource "kvindo_open_vpn_user_settings" "example" {
  metadata = {
    name = "restricted-settings"
  }
  spec = {
    open_vpn_id        = kvindo_open_vpn.main.id
    allowed_ipv4_cidrs = ["10.0.0.0/8", "192.168.0.0/16"]
    denied_domains     = ["social-media.example.com"]
  }
}
