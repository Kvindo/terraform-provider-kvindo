resource "kvindo_vpc" "main" {
  metadata = {
    name = "my-vpc"
  }
  spec = {
    hosting_provider_id = "01abc123def456gh789012345"
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

resource "kvindo_ollama" "example" {
  metadata = {
    name = "my-ollama"
  }
  spec = {
    vpc_subnet_id   = kvindo_vpc_subnet.main.id
    vm_offer_id     = "01vm0ffr123456789012345"
    models          = ["llama3.1:8b", "qwen2.5:7b"]
    root_password   = var.ollama_password
    volume_size_gib = 100
  }
}

variable "ollama_password" {
  type      = string
  sensitive = true
}

output "ollama_host" {
  value = kvindo_ollama.example.status.host
}
