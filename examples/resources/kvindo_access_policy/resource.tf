resource "kvindo_access_policy" "example" {
  metadata = {
    name = "read-only-policy"
  }
  spec = {
    content = jsonencode({
      rules = [
        {
          action    = "read"
          resources = ["*"]
        }
      ]
    })
  }
}
