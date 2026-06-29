resource "kvindo_vpc" "main" {
  metadata = {
    name = "my-vpc"
  }
  spec = {
    hosting_provider_id = "01abc123def456gh789012345"
    ipv4_cidr           = "10.0.0.0/16"
  }
}

resource "kvindo_vpc_subnet" "main" {
  metadata = {
    name = "my-subnet"
  }
  spec = {
    vpc_id    = kvindo_vpc.main.id
    ipv4_cidr = "10.0.1.0/24"
  }
}

resource "kvindo_postgresql_standalone" "example" {
  metadata = {
    name = "my-postgres"
  }
  spec = {
    version               = "16"
    vpc_subnet_id         = kvindo_vpc_subnet.main.id
    vm_offer_id           = "01vm0ffr123456789012345"
    root_password         = var.db_password
    volume_offer_id       = "01vol0ffr12345678901234"
    volume_size_gib       = 50
    backup_retention_days = 7
  }
}

variable "db_password" {
  type      = string
  sensitive = true
}
