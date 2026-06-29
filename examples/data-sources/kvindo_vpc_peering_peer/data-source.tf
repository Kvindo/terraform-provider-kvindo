data "kvindo_vpc_peering_peer" "example" {
  id = "01abc123def456gh789012345"
}

output "info_public_ip_v4" {
  value = data.kvindo_vpc_peering_peer.example.status.public_ip_v4
}
