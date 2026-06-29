data "kvindo_vpc_peering_external_peer" "example" {
  id = "01abc123def456gh789012345"
}

output "info_state" {
  value = data.kvindo_vpc_peering_external_peer.example.status.state
}
