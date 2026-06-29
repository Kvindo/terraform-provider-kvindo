data "kvindo_image" "example" {
  id = "01abc123def456gh789012345"
}

output "info_size_bytes" {
  value = data.kvindo_image.example.status.size_bytes
}
