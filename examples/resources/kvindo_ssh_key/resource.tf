resource "kvindo_ssh_key" "example" {
  metadata = {
    name = "my-ssh-key"
  }
  spec = {
    public_key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIHKGbFGS... user@host"
  }
}
