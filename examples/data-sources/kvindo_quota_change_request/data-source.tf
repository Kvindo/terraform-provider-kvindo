data "kvindo_quota_change_request" "example" {
  id = "01abc123def456gh789012345"
}

output "info_ticket_id" {
  value = data.kvindo_quota_change_request.example.status.ticket_id
}
