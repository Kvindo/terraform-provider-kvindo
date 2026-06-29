data "kvindo_volume_attachment" "example" {
  id = "01abc123def456gh789012345"
}

output "info_state" {
  value = data.kvindo_volume_attachment.example.status.state
}
