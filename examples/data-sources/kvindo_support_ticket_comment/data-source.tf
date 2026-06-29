data "kvindo_support_ticket_comment" "example" {
  id = "01abc123def456gh789012345"
}

output "info_state" {
  value = data.kvindo_support_ticket_comment.example.status.state
}
