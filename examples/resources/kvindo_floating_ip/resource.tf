resource "kvindo_floating_ip" "example" {
  metadata = {
    name = "my-floating-ip"
  }
  spec = {
    hosting_provider_id = "01abc123def456gh789012345"
  }
}

output "public_ip" {
  value = kvindo_floating_ip.example.status.public_ip_v4
}
