data "kvindo_kafka_parameters_set" "durable" {
  name = "durable"
}

output "parameters" {
  value = data.kvindo_kafka_parameters_set.durable.spec.parameters
}
