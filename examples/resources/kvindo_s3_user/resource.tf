resource "kvindo_s3_bucket" "main" {
  metadata = {
    name = "my-app-bucket"
  }
  spec = {
    tier   = "standard"
    region = "ru-msk-1"
  }
}

resource "kvindo_s3_user_access_policy" "rw" {
  metadata = {
    name = "readwrite-policy"
  }
  spec = {
    policy_json = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Effect   = "Allow"
          Action   = ["s3:GetObject", "s3:PutObject", "s3:DeleteObject", "s3:ListBucket"]
          Resource = ["arn:aws:s3:::${kvindo_s3_bucket.main.name}/*"]
        }
      ]
    })
  }
}

resource "kvindo_s3_user" "example" {
  metadata = {
    name = "app-user"
  }
  spec = {
    bucket_id         = kvindo_s3_bucket.main.id
    access_policy_ids = [kvindo_s3_user_access_policy.rw.id]
  }
}

output "access_key" {
  value = kvindo_s3_user.example.status.access_key
}

output "secret_key" {
  value     = kvindo_s3_user.example.status.secret_key
  sensitive = true
}
