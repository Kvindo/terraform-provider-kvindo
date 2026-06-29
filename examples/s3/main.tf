terraform {
  required_version = ">= 1.0"

  required_providers {
    kvindo = {
      source  = "registry.terraform.io/kvindo/kvindo"
      version = "~> 1.0"
    }
  }
}

provider "kvindo" {
  token = var.kvindo_token
  # endpoint = "https://cloud-api.kvindo.com"  # optional, this is the default
}

# ── Variables ────────────────────────────────────────────────────────────────

variable "kvindo_token" {
  description = "Kvindo Cloud API token"
  type        = string
  sensitive   = true
}

variable "folder_id" {
  description = "Folder ID to create resources in"
  type        = string
}

# ── S3 Bucket ────────────────────────────────────────────────────────────────

resource "kvindo_s3_bucket" "main" {
  name      = "my-app-bucket"
  folder_id = var.folder_id

  tier         = "standard"
  region       = "ru-msk-1"
  is_public    = false
  is_versioned = true
  quota_gib    = 100

  labels = {
    env = "production"
    app = "my-app"
  }
}

# ── Access Policy ─────────────────────────────────────────────────────────────

resource "kvindo_s3_user_access_policy" "readwrite" {
  name      = "my-app-readwrite"
  folder_id = var.folder_id

  policy_json = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "ReadWrite"
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:PutObject",
          "s3:DeleteObject",
          "s3:ListBucket",
        ]
        Resource = [
          "arn:aws:s3:::${kvindo_s3_bucket.main.name}",
          "arn:aws:s3:::${kvindo_s3_bucket.main.name}/*",
        ]
      },
    ]
  })
}

resource "kvindo_s3_user_access_policy" "readonly" {
  name      = "my-app-readonly"
  folder_id = var.folder_id

  policy_json = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "ReadOnly"
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:ListBucket",
        ]
        Resource = [
          "arn:aws:s3:::${kvindo_s3_bucket.main.name}",
          "arn:aws:s3:::${kvindo_s3_bucket.main.name}/*",
        ]
      },
    ]
  })
}

# ── S3 Users ──────────────────────────────────────────────────────────────────

resource "kvindo_s3_user" "app" {
  name      = "my-app-user"
  folder_id = var.folder_id
  bucket_id = kvindo_s3_bucket.main.id

  access_policy_ids = [kvindo_s3_user_access_policy.readwrite.id]
}

resource "kvindo_s3_user" "ci" {
  name      = "my-app-ci-user"
  folder_id = var.folder_id
  bucket_id = kvindo_s3_bucket.main.id

  access_policy_ids = [kvindo_s3_user_access_policy.readonly.id]
}

# ── Outputs ───────────────────────────────────────────────────────────────────

output "bucket_endpoint" {
  description = "S3 endpoint URL for the bucket"
  value       = kvindo_s3_bucket.main.info.endpoint_url
}

output "app_user_access_key" {
  description = "Access key for the app S3 user"
  value       = kvindo_s3_user.app.info.access_key
  sensitive   = true
}

output "app_user_secret_key" {
  description = "Secret key for the app S3 user"
  value       = kvindo_s3_user.app.info.secret_key
  sensitive   = true
}

output "ci_user_access_key" {
  description = "Access key for the CI S3 user (read-only)"
  value       = kvindo_s3_user.ci.info.access_key
  sensitive   = true
}

output "ci_user_secret_key" {
  description = "Secret key for the CI S3 user (read-only)"
  value       = kvindo_s3_user.ci.info.secret_key
  sensitive   = true
}
