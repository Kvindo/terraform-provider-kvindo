resource "kvindo_s3_bucket" "main" {
  metadata = {
    name = "my-app-bucket"
  }
  spec = {
    tier   = "standard"
    region = "ru-msk-1"
  }
}

resource "kvindo_s3_user_access_policy" "example" {
  metadata = {
    name = "readwrite-policy"
  }
  spec = {
    policy_json = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Sid    = "ReadWrite"
          Effect = "Allow"
          Action = ["s3:GetObject", "s3:PutObject", "s3:DeleteObject", "s3:ListBucket"]
          Resource = [
            "arn:aws:s3:::${kvindo_s3_bucket.main.name}",
            "arn:aws:s3:::${kvindo_s3_bucket.main.name}/*",
          ]
        }
      ]
    })
  }
}
