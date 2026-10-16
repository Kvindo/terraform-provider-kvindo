data "kvindo_opensearch" "example" {
  name = "example"
}

output "example_id" {
  value = data.kvindo_opensearch.example.metadata.id
}
