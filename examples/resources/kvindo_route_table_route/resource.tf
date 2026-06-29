resource "kvindo_route_table" "main" {
  metadata = {
    name = "my-route-table"
  }
}

resource "kvindo_route_table_route" "example" {
  metadata = {
    name = "internet-route"
  }
  spec = {
    route_table_id   = kvindo_route_table.main.id
    destination_cidr = "0.0.0.0/0"
    target_ip        = "10.0.0.1"
  }
}
