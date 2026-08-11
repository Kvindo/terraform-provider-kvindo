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
    # vpc_id and vpc_subnet_id are mutually exclusive: vpc_id attaches the route table to the
    # whole VPC (routes apply VPC-wide); vpc_subnet_id attaches it to a single subnet only
    # (routes are policy-routed so they apply solely to that subnet's traffic). Set exactly one.
    vpc_id = kvindo_vpc.main.id
    # vpc_subnet_id = kvindo_vpc_subnet.main.id
  }
}
