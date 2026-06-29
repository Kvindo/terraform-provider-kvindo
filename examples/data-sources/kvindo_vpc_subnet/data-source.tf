data "kvindo_vpc_subnet" "example" {
  id = "01abc123def456gh789012345"
}

output "info_state" {
  value = data.kvindo_vpc_subnet.example.status.state
}
