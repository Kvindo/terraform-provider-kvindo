data "kvindo_vpc" "example" {
  id = "01abc123def456gh789012345"
}

output "info_nat_public_ip_v4" {
  value = data.kvindo_vpc.example.status.nat_public_ip_v4
}
