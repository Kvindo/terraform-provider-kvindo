data "kvindo_kafka_user" "app" {
  name = "app-user"
}

output "enabled" {
  value = data.kvindo_kafka_user.app.spec.enabled
}
