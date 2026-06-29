# A transaction creates multiple resources atomically — all succeed or all roll back.
# Each sub-resource type is a map keyed by a name you choose (the key is preserved in
# state across applies). Cross-reference one sub-resource from another by pre-assigning
# its id under metadata.id (a ULID you generate) and referencing that same value where
# it's needed — here the s3_user points at the bucket and policy created in the same call.
locals {
  bucket_id = "01j8z0bckt0000000000000000"
  policy_id = "01j8z0pcy00000000000000000"
}

resource "kvindo_transaction" "example" {
  metadata = {
    name = "bootstrap"
  }

  spec = {
    delete_resources_on_transaction_delete = true

    folders = {
      "prod" = {
        metadata = { name = "production", description = "Production resources" }
      }
    }

    ssh_keys = {
      "deploy" = {
        metadata = { name = "deploy-key" }
        spec     = { public_key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIHKGbFGS... user@host" }
      }
    }

    s3_buckets = {
      "assets" = {
        metadata = { id = local.bucket_id, name = "app-assets" }
        spec     = { tier = "standard", region = "ru-msk-1", quota_gib = 100 }
      }
    }

    s3_user_access_policies = {
      "rw" = {
        metadata = { id = local.policy_id, name = "assets-rw" }
        spec = {
          policy_json = jsonencode({
            Version   = "2012-10-17"
            Statement = [{ Effect = "Allow", Action = ["s3:*"], Resource = ["*"] }]
          })
        }
      }
    }

    s3_users = {
      "app" = {
        metadata = { name = "app-user" }
        spec = {
          bucket_id         = local.bucket_id   # cross-reference the bucket above
          access_policy_ids = [local.policy_id] # cross-reference the policy above
        }
      }
    }
  }
}

output "access_key" {
  value     = kvindo_transaction.example.spec.s3_users["app"].status.access_key
  sensitive = true
}
