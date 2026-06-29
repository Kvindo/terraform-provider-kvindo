data "kvindo_support_ticket_comment_attachment" "example" {
  id = "01abc123def456gh789012345"
}

output "info_download_url" {
  value = data.kvindo_support_ticket_comment_attachment.example.status.download_url
}
