provider "kvindo" {
  token = var.kvindo_token
  # endpoint = "https://cloud-api.kvindo.com"  # optional, this is the default
}

variable "kvindo_token" {
  description = "Kvindo Cloud API token"
  type        = string
  sensitive   = true
}
