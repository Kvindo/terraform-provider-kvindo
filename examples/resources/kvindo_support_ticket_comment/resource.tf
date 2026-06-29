resource "kvindo_support_ticket" "main" {
  metadata = {
    name = "my-ticket"
  }
  spec = {
    kind     = "technical"
    severity = "medium"
  }
}

resource "kvindo_support_ticket_comment" "example" {
  metadata = {
    name = "initial-comment"
  }
  spec = {
    ticket_id = kvindo_support_ticket.main.id
    content   = "The issue started after the last maintenance window."
  }
}
