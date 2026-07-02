resource "kvindo_vpc" "main" {
  metadata = {
    name = "my-vpc"
  }
  spec = {
    hosting_provider_id = "01abc123def456gh789012345"
  }
}

resource "kvindo_route_table" "main" {
  metadata = {
    name = "my-route-table"
  }
}

resource "kvindo_route_table_attachment" "example" {
  metadata = {
    name = "my-rt-attachment"
  }
  spec = {
    route_table_id = kvindo_route_table.main.id
    vpc_id         = kvindo_vpc.main.id
  }
}
