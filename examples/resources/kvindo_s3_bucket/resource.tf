resource "kvindo_s3_bucket" "example" {
  metadata = {
    name = "my-app-bucket"
  }
  spec = {
    tier         = "standard"
    region       = "ru-msk-1"
    is_versioned = true
    is_public    = false
    quota_gib    = 100
  }
}

output "endpoint_url" {
  value = kvindo_s3_bucket.example.status.endpoint_url
}
