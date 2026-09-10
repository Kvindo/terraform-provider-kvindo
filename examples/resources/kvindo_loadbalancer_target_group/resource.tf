# A target group has no spec of its own - it's a standalone bucket of backends, associated with a
# loadbalancer later by referencing its id from a listener rule's forward action (see
# kvindo_loadbalancer_http_listener_rule's example), not by any field on the target group itself.
resource "kvindo_loadbalancer_target_group" "example" {
  metadata = {
    name = "my-target-group"
  }
}
