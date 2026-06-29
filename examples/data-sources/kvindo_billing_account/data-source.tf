data "kvindo_billing_account" "example" {
  id = "01abc123def456gh789012345"
}

output "info_rub_balance" {
  value = data.kvindo_billing_account.example.status.rub_balance
}
