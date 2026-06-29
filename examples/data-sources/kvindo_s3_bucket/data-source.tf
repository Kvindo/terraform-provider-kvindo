data "kvindo_s3_bucket" "example" {
  id = "01abc123def456gh789012345"
}

output "info_endpoint_url" {
  value = data.kvindo_s3_bucket.example.status.endpoint_url
}
