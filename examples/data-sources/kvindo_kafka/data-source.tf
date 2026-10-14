data "kvindo_kafka" "main" {
  name = "main"
}

output "bootstrap_servers" {
  value = data.kvindo_kafka.main.status.bootstrap_servers_plaintext
}
