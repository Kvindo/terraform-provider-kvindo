resource "kvindo_support_ticket" "main" {
  metadata = {
    name = "my-ticket"
  }
  spec = {
    kind     = "technical"
    severity = "medium"
  }
}

resource "kvindo_support_ticket_comment" "main" {
  metadata = {
    name = "initial-comment"
  }
  spec = {
    ticket_id = kvindo_support_ticket.main.id
    content   = "Attaching logs for reference."
  }
}

resource "kvindo_support_ticket_comment_attachment" "example" {
  metadata = {
    name = "app-logs"
  }
  spec = {
    comment_id          = kvindo_support_ticket_comment.main.id
    file_name           = "app.log"
    file_type           = "text/plain"
    file_content_base64 = filebase64("app.log")
  }
}
