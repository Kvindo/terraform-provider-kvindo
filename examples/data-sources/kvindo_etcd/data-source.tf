data "kvindo_etcd" "example" {
  id = "01abc123def456gh789012345"
}

output "etcd_state" {
  value = data.kvindo_etcd.example.status.state
}
