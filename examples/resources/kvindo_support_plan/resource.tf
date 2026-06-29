resource "kvindo_support_plan" "example" {
  metadata = {
    name = "business-support"
  }
  spec = {
    tier = "business"
  }
}
