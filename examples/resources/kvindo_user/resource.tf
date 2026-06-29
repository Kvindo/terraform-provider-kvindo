resource "kvindo_access_policy" "viewer" {
  metadata = {
    name = "viewer-policy"
  }
  spec = {
    content = jsonencode({
      rules = [{ action = "read", resources = ["*"] }]
    })
  }
}

resource "kvindo_user" "example" {
  metadata = {
    name = "alice"
  }
  spec = {
    email             = "alice@example.com"
    access_policy_ids = [kvindo_access_policy.viewer.id]
  }
}
