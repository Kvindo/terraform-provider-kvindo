resource "kvindo_support_ticket" "example" {
  metadata = {
    name = "vm-connectivity-issue"
  }
  spec = {
    kind     = "technical"
    severity = "medium"
  }
}
