data "kvindo_opensearch_parameters_set" "example" {
  name = "example"
}

output "example_id" {
  value = data.kvindo_opensearch_parameters_set.example.metadata.id
}
