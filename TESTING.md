# Testing the Kvindo Terraform Provider

## Prerequisites

- Go 1.21 or later
- Terraform 1.5 or later
- A Kvindo Cloud API token

## Building the Provider

```bash
cd terraform-provider-kvindo
go build -o terraform-provider-kvindo .
```

Or using make:
```bash
make build
```

The binary `terraform-provider-kvindo` will be created in the current directory.

## Configuring Terraform to Use the Local Provider

Create or update `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "registry.terraform.io/kvindo/kvindo" = "/path/to/terraform-provider-kvindo/directory"
  }
  direct {}
}
```

Replace `/path/to/terraform-provider-kvindo/directory` with the absolute path to the directory containing the built binary.

**Important**: When using `dev_overrides`, Terraform skips the provider locking step, so you don't need to run `terraform init` to install the provider. However, you still need `terraform init` to initialize other things (backend, modules, etc.).

## Example Configuration

Create a `main.tf` file:

```hcl
terraform {
  required_providers {
    kvindo = {
      source = "registry.terraform.io/kvindo/kvindo"
    }
  }
}

provider "kvindo" {
  token    = "your-api-token-here"
  endpoint = "https://cloud-api.kvindo.com"  # optional, this is the default
}

# Create a VPC
resource "kvindo_vpc" "example" {
  name                = "my-vpc"
  hosting_provider_id = "provider-id"
  ipv4_cidr           = "10.0.0.0/16"
  folder_id           = "folder-id"
  description         = "My example VPC"
}

# Create a subnet
resource "kvindo_vpc_subnet" "example" {
  name      = "my-subnet"
  vpc_id    = kvindo_vpc.example.id
  ipv4_cidr = "10.0.1.0/24"
}

# Create a VM
resource "kvindo_vm" "example" {
  name           = "my-vm"
  offer_id       = "vm-offer-id"
  image_id       = "image-id"
  vpc_subnet_id  = kvindo_vpc_subnet.example.id
  folder_id      = "folder-id"
}

# Read data from an existing VPC
data "kvindo_vpc" "existing" {
  id = "existing-vpc-id"
}

output "vpc_id" {
  value = kvindo_vpc.example.id
}

output "vm_private_ip" {
  value = kvindo_vm.example.info_private_ipv4
}
```

## Running Terraform Commands

Initialize Terraform (sets up backend and modules but skips provider installation when using dev_overrides):
```bash
terraform init
```

Preview changes:
```bash
terraform plan
```

Apply changes:
```bash
terraform apply
```

Destroy resources:
```bash
terraform destroy
```

## Importing Existing Resources

All resources support import. Use the resource's API ID:

```bash
terraform import kvindo_vpc.example <vpc-uuid>
terraform import kvindo_vm.example <vm-uuid>
```

## Using Environment Variables for Authentication

Instead of hardcoding the token in `main.tf`, use environment variables:

```bash
export KVINDO_TOKEN="your-api-token-here"
export KVINDO_ENDPOINT="https://cloud-api.kvindo.com"  # optional
```

Then the provider block can be:
```hcl
provider "kvindo" {}
```

## Regenerating from Latest API Spec

To regenerate all resource files from the latest Kvindo API swagger spec:

```bash
make regenerate
```

Or manually:
```bash
# Download latest swagger
curl -f "https://cloud-api.kvindo.com/swagger/v1/swagger.json" -o kvindo-api.json

# Run generator
cd tools/generator
go run . --swagger ../../kvindo-api.json --output ../../internal/provider

# Rebuild
cd ../..
go build ./...
```

## Resource Reference

The provider implements the following resources and data sources:

### Compute
- `kvindo_vm` / `data.kvindo_vm`
- `kvindo_volume` / `data.kvindo_volume`
- `kvindo_volume_attachment` / `data.kvindo_volume_attachment`
- `kvindo_image` / `data.kvindo_image`
- `kvindo_image_schedule` / `data.kvindo_image_schedule`
- `kvindo_ssh_key` / `data.kvindo_ssh_key`
- `kvindo_ssh_private_key` / `data.kvindo_ssh_private_key`
- `kvindo_certificate` / `data.kvindo_certificate`
- `kvindo_vm_on_off_maintenance_action` / `data.kvindo_vm_on_off_maintenance_action`
- `kvindo_vm_recurrent_command_maintenance_action` / `data.kvindo_vm_recurrent_command_maintenance_action`

### Networking
- `kvindo_vpc` / `data.kvindo_vpc`
- `kvindo_vpc_subnet` / `data.kvindo_vpc_subnet`
- `kvindo_floating_ip` / `data.kvindo_floating_ip`
- `kvindo_security_group` / `data.kvindo_security_group`
- `kvindo_nat_gateway` / `data.kvindo_nat_gateway`
- `kvindo_route_table` / `data.kvindo_route_table`
- `kvindo_route_table_route` / `data.kvindo_route_table_route`
- `kvindo_route_table_attachment` / `data.kvindo_route_table_attachment`
- `kvindo_vpc_peering` / `data.kvindo_vpc_peering`
- `kvindo_vpc_peering_peer` / `data.kvindo_vpc_peering_peer`
- `kvindo_vpc_peering_external_peer` / `data.kvindo_vpc_peering_external_peer`

### IAM / Organization
- `kvindo_folder` / `data.kvindo_folder`
- `kvindo_user` / `data.kvindo_user`
- `kvindo_user_token` / `data.kvindo_user_token`
- `kvindo_access_policy` / `data.kvindo_access_policy`
- `kvindo_billing_account` / `data.kvindo_billing_account`
- `kvindo_quota` / `data.kvindo_quota`
- `kvindo_quota_change_request` / `data.kvindo_quota_change_request`
- `kvindo_hosting_provider` / `data.kvindo_hosting_provider`

### Kubernetes
- `kvindo_kubernetes` / `data.kvindo_kubernetes`
- `kvindo_kubernetes_node_group` / `data.kvindo_kubernetes_node_group`
- `kvindo_kubernetes_user` / `data.kvindo_kubernetes_user`
- `kvindo_kubernetes_user_role` / `data.kvindo_kubernetes_user_role`

### Load Balancer
- `kvindo_loadbalancer` / `data.kvindo_loadbalancer`
- `kvindo_loadbalancer_target_group` / `data.kvindo_loadbalancer_target_group`
- `kvindo_loadbalancer_target_group_static_target` / `data.kvindo_loadbalancer_target_group_static_target`
- `kvindo_loadbalancer_target_group_service_discovery_target` / `data.kvindo_loadbalancer_target_group_service_discovery_target`
- `kvindo_loadbalancer_http_listener` / `data.kvindo_loadbalancer_http_listener`
- `kvindo_loadbalancer_https_listener` / `data.kvindo_loadbalancer_https_listener`
- `kvindo_loadbalancer_http_listener_rule` / `data.kvindo_loadbalancer_http_listener_rule`
- `kvindo_loadbalancer_https_listener_rule` / `data.kvindo_loadbalancer_https_listener_rule`
- `kvindo_loadbalancer_tcp_listener` / `data.kvindo_loadbalancer_tcp_listener`
- `kvindo_loadbalancer_tls_listener` / `data.kvindo_loadbalancer_tls_listener`
- `kvindo_loadbalancer_tcp_listener_rule` / `data.kvindo_loadbalancer_tcp_listener_rule`
- `kvindo_loadbalancer_tls_listener_rule` / `data.kvindo_loadbalancer_tls_listener_rule`
- `kvindo_loadbalancer_udp_listener` / `data.kvindo_loadbalancer_udp_listener`
- `kvindo_loadbalancer_udp_listener_rule` / `data.kvindo_loadbalancer_udp_listener_rule`

### Databases
- `kvindo_postgresql` / `data.kvindo_postgresql`
- `kvindo_postgresql_node_group` / `data.kvindo_postgresql_node_group`
- `kvindo_postgresql_parameters_set` / `data.kvindo_postgresql_parameters_set`
- `kvindo_postgresql_standalone` / `data.kvindo_postgresql_standalone`
- `kvindo_etcd` / `data.kvindo_etcd`
- `kvindo_etcd_node_group` / `data.kvindo_etcd_node_group`

### Object Storage
- `kvindo_s3_bucket` / `data.kvindo_s3_bucket`
- `kvindo_s3_user` / `data.kvindo_s3_user`
- `kvindo_s3_user_access_policy` / `data.kvindo_s3_user_access_policy`

### Monitoring
- `kvindo_victoria_metrics` / `data.kvindo_victoria_metrics`
- `kvindo_grafana` / `data.kvindo_grafana`
- `kvindo_ollama` / `data.kvindo_ollama`

### VPN
- `kvindo_open_vpn` / `data.kvindo_open_vpn`
- `kvindo_open_vpn_user` / `data.kvindo_open_vpn_user`
- `kvindo_open_vpn_user_settings` / `data.kvindo_open_vpn_user_settings`

### Dev Tools
- `kvindo_gitlab` / `data.kvindo_gitlab`
- `kvindo_gitlab_runner` / `data.kvindo_gitlab_runner`

### Support
- `kvindo_support_plan` / `data.kvindo_support_plan`
- `kvindo_support_ticket` / `data.kvindo_support_ticket`
- `kvindo_support_ticket_comment` / `data.kvindo_support_ticket_comment`
- `kvindo_support_ticket_comment_attachment` / `data.kvindo_support_ticket_comment_attachment`

## Notes on Async Operations

All create, update, and delete operations are asynchronous. The provider automatically polls the request status endpoint until the operation completes (with a 10-minute timeout and exponential backoff starting at 2 seconds).

## Sensitive Fields

The following fields are marked as sensitive and will not be shown in plan output:
- Passwords (`root_password`, etc.)
- Tokens (`info_token`, `info_kubeconfig`)
- Private keys (`private_key`, `private_key_pem`, `info_client_key_pem`)
- Secrets (`info_secret_key`, `info_access_key`)
- Certificates (`certificate_pem`, `info_ca_certificate_pem`, etc.)
- VPN config (`info_config`)
