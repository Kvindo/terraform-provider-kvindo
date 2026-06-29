resource "kvindo_kubernetes_user_role" "example" {
  metadata = {
    name = "pod-viewer"
  }
  spec = {
    api_groups = [""]
    resources  = ["pods", "pods/log", "services"]
    verbs      = ["get", "list", "watch"]
    namespaces = ["default", "production"]
  }
}
