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

resource "kvindo_gitlab_runner" "example" {
  metadata = {
    name = "my-runner"
  }
  spec = {
    vpc_subnet_id = kvindo_vpc_subnet.main.id
    vm_offer_id   = "01vm0ffr123456789012345"
    concurrency   = 4
  }
}
