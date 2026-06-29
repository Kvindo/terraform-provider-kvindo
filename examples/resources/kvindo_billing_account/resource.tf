resource "kvindo_billing_account" "example" {
  metadata = {
    name = "my-billing-account"
  }
}

output "balance" {
  value = kvindo_billing_account.example.status.rub_balance
}
