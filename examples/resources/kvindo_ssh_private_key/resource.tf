resource "kvindo_ssh_private_key" "example" {
  metadata = {
    name = "my-deploy-key"
  }
  spec = {
    private_key = file("~/.ssh/id_rsa")
  }
}
