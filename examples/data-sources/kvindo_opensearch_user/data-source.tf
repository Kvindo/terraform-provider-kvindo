data "kvindo_opensearch_user" "example" {
  name = "example"
}

output "example_id" {
  value = data.kvindo_opensearch_user.example.metadata.id
}
