resource "kvindo_user" "main" {
  metadata = {
    name = "alice"
  }
  spec = {
    email = "alice@example.com"
  }
}

resource "kvindo_user_token" "example" {
  metadata = {
    name = "ci-token"
  }
  spec = {
    user_id       = kvindo_user.main.id
    send_to_email = false
  }
}

output "token" {
  value     = kvindo_user_token.example.status.token
  sensitive = true
}
