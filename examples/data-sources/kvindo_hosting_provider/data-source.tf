# Look up a hosting provider by name (or by id instead — exactly one is required).
data "kvindo_hosting_provider" "example" {
  name = "ru-msk-1"
}

output "hosting_provider_id" {
  value = data.kvindo_hosting_provider.example.metadata.id
}
