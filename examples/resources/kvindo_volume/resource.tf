resource "kvindo_volume" "example" {
  metadata = {
    name = "my-volume"
  }
  spec = {
    hosting_provider_id = "01abc123def456gh789012345"
    offer_id            = "01vol0ffr12345678901234"
    size_gib            = 50
  }
}
