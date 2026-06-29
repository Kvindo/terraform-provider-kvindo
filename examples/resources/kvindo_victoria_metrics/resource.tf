resource "kvindo_vpc" "main" {
  metadata = {
    name = "my-vpc"
  }
  spec = {
    hosting_provider_id = "01abc123def456gh789012345"
    ipv4_cidr           = "10.0.0.0/16"
  }
}

resource "kvindo_victoria_metrics" "example" {
  metadata = {
    name = "my-metrics"
  }
  spec = {
    vpc_id             = kvindo_vpc.main.id
    create_public_ipv4 = true
  }
}

output "fqdn" {
  value = kvindo_victoria_metrics.example.status.fqdn
}
