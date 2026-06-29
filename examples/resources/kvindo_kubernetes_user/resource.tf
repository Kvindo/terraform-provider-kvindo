resource "kvindo_kubernetes" "main" {
  metadata = {
    name = "my-cluster"
  }
  spec = {
    version = "1.30"
  }
}

resource "kvindo_kubernetes_user_role" "viewer" {
  metadata = {
    name = "viewer-role"
  }
  spec = {
    api_groups = [""]
    resources  = ["pods", "services"]
    verbs      = ["get", "list", "watch"]
    namespaces = ["default"]
  }
}

resource "kvindo_kubernetes_user" "example" {
  metadata = {
    name = "alice"
  }
  spec = {
    kubernetes_id = kvindo_kubernetes.main.id
    role_ids      = [kvindo_kubernetes_user_role.viewer.id]
  }
}

output "kubeconfig" {
  value     = kvindo_kubernetes_user.example.status.kubeconfig
  sensitive = true
}
