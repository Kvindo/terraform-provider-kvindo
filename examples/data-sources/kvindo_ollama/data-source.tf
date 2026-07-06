data "kvindo_ollama" "example" {
  id = "01abc123def456gh789012345"
}

output "ollama_host" {
  value = data.kvindo_ollama.example.status.host
}
