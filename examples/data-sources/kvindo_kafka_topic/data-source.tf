data "kvindo_kafka_topic" "orders" {
  name = "orders"
}

output "partitions" {
  value = data.kvindo_kafka_topic.orders.spec.partitions
}
