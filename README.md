# Terraform Provider for Kvindo Cloud

Manages resources on [Kvindo Cloud](https://cloud.kvindo.com) — a multi-tenant cloud platform for VMs, object storage, Kubernetes, load balancers, VPNs, and managed services.

## Requirements

- [Terraform](https://www.terraform.io/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21 (to build from source)

## Building

```bash
cd terraform-provider-kvindo
go build ./...
```

The `~/.terraformrc` dev override points to this directory, so after rebuilding, `terraform plan/apply` picks up the new binary automatically.

## Configuration

```hcl
terraform {
  required_providers {
    kvindo = {
      source = "registry.terraform.io/kvindo/kvindo"
    }
  }
}

provider "kvindo" {
  token    = var.kvindo_token
  endpoint = "https://cloud-api.kvindo.com"
}
```

The token is a long-lived JWT from the Kvindo Cloud portal (`/iam/user-tokens`).

### Field Naming

Resource types and attribute names mirror the Kvindo Cloud REST API. Every field
corresponds to the request/response schema published in the OpenAPI (Swagger) spec at
<https://cloud-api.kvindo.com/swagger> — that spec is the source of truth for field
names, types, and accepted values. The provider only rewrites casing: the API's
`camelCase` becomes Terraform `snake_case` (e.g. API `volumeSizeGiB` → `volume_size_gib`,
`floatingIpId` → `floating_ip_id`). Read-only fields the API returns are grouped under a
single computed [`info` block](#the-info-block) rather than scattered across the schema.

## Resource Categories

### Compute
- `kvindo_vm` — Virtual machine
- `kvindo_volume` / `kvindo_volume_attachment` — Block storage
- `kvindo_image` / `kvindo_image_schedule` — Custom images and snapshots
- `kvindo_ssh_key` / `kvindo_ssh_private_key` / `kvindo_certificate`

### Networking
- `kvindo_vpc` / `kvindo_vpc_subnet`
- `kvindo_floating_ip` / `kvindo_security_group`
- `kvindo_route_table` / `kvindo_route_table_route` / `kvindo_route_table_attachment`
- `kvindo_vpc_peering` / `kvindo_vpc_peering_peer` / `kvindo_vpc_peering_external_peer`

### Load Balancer
- `kvindo_loadbalancer`
- `kvindo_loadbalancer_target_group` / `_static_target` / `_service_discovery_target`
- Listeners: `_http_listener`, `_https_listener`, `_tcp_listener`, `_tls_listener`, `_udp_listener`
- Rules: `_http_listener_rule`, `_https_listener_rule`, `_tcp_listener_rule`, `_tls_listener_rule`, `_udp_listener_rule`

### Object Storage
- `kvindo_s3_bucket` — S3-compatible bucket
- `kvindo_s3_user` — S3 IAM user
- `kvindo_s3_user_access_policy` — S3 IAM policy (JSON)

### Kubernetes
- `kvindo_kubernetes` / `kvindo_kubernetes_node_group`
- `kvindo_kubernetes_user` / `kvindo_kubernetes_user_role`

### Databases
- `kvindo_postgresql_standalone` — PostgreSQL Standalone instance
- `kvindo_postgresql_parameters_set` — PostgreSQL parameter configuration

### VPN
- `kvindo_open_vpn` / `kvindo_open_vpn_user` / `kvindo_open_vpn_user_settings`

### Monitoring
- `kvindo_victoria_metrics`

### Dev Tools
- `kvindo_gitlab` / `kvindo_gitlab_runner`

### IAM / Organization
- `kvindo_folder` — Resource namespace
- `kvindo_user` / `kvindo_user_token`
- `kvindo_access_policy` / `kvindo_billing_account`
- `kvindo_quota` / `kvindo_quota_change_request`
- `kvindo_hosting_provider`

### Support
- `kvindo_support_plan` / `kvindo_support_ticket` / `kvindo_support_ticket_comment` / `kvindo_support_ticket_comment_attachment`

### Transaction (atomic multi-resource)
- `kvindo_transaction` — Creates multiple sub-resources in a single API call. Sub-resource types: `folders`, `ssh_keys`, `s3_buckets`, `s3_user_access_policies`, `s3_users` (all map attributes, keyed by user-chosen name).

## The `info` Block

Every resource exposes a computed `info` nested attribute with read-only fields returned by the API:

```hcl
output "state"    { value = kvindo_folder.main.info.state }
output "endpoint" { value = kvindo_s3_bucket.main.info.endpoint_url }
output "key"      { value = kvindo_s3_user.app.info.access_key  sensitive = true }
```

| Resource | `info` fields |
|---|---|
| Most resources | `state` |
| `kvindo_s3_bucket` | `state`, `endpoint_url` |
| `kvindo_s3_user` | `state`, `access_key`, `secret_key` |
| `kvindo_postgresql_standalone` | `state`, `root_user_name`, `public_ip_v4`, `private_ip_v4`, `port` |
| `kvindo_vm` | `state`, `private_ipv4`, `public_ipv4`, `private_ipv6`, `public_ipv6` |
| `kvindo_floating_ip` | `state`, `public_ip_v4` |
| `kvindo_loadbalancer`, `kvindo_vpc_peering_peer` | `state`, `public_ip_v4`, `public_ip_v6`, `private_ip_v4`, `private_ip_v6` |
| `kvindo_victoria_metrics`, `kvindo_gitlab` | `state`, `public_ip_v4`, `public_ip_v6`, `private_ip_v4`, `private_ip_v6`, `fqdn` |
| `kvindo_kubernetes` | `state`, `api_server_url` |
| `kvindo_kubernetes_user` | `state`, `kubeconfig` |
| `kvindo_open_vpn_user` | `state`, `config` |
| `kvindo_user_token` | `state`, `token` |
| `kvindo_image` | `state`, `size_bytes` |
| `kvindo_vpc` | `state`, `nat_public_ip_v4` |
| `kvindo_billing_account` | `state`, `rub_balance` |
| `kvindo_quota` | `state`, `current_value` |
| `kvindo_quota_change_request` | `state`, `ticket_id` |

Transaction sub-resources expose the same `info` fields as their standalone counterparts. Access via `kvindo_transaction.main.s3_users["app"].info.access_key`.

## Example: Folder + S3 Bucket + Users

```hcl
resource "random_id" "suffix" { byte_length = 4 }

resource "kvindo_folder" "main" {
  name = "my-app"
}

resource "kvindo_s3_bucket" "main" {
  name      = "my-app-${random_id.suffix.hex}"
  folder_id = kvindo_folder.main.id
  tier      = "standard"
  region    = "ru-msk-1"
  quota_gib = 10
}

resource "kvindo_s3_user_access_policy" "rw" {
  name        = "my-app-rw-${random_id.suffix.hex}"
  folder_id   = kvindo_folder.main.id
  policy_json = jsonencode({
    Version = "2012-10-17"
    Statement = [{ Effect = "Allow", Action = ["s3:*"], Resource = ["arn:aws:s3:::${kvindo_s3_bucket.main.name}/*"] }]
  })
}

resource "kvindo_s3_user" "app" {
  name              = "my-app-user-${random_id.suffix.hex}"
  folder_id         = kvindo_folder.main.id
  bucket_id         = kvindo_s3_bucket.main.id
  access_policy_ids = [kvindo_s3_user_access_policy.rw.id]
}

output "endpoint"   { value = kvindo_s3_bucket.main.info.endpoint_url }
output "access_key" { value = kvindo_s3_user.app.info.access_key  sensitive = true }
output "secret_key" { value = kvindo_s3_user.app.info.secret_key  sensitive = true }
```

## Example: Atomic Transaction

Creates bucket + policies + users in a single API round-trip:

```hcl
resource "kvindo_transaction" "main" {
  name      = "my-app"
  folder_id = kvindo_folder.main.id
  delete_resources_on_transaction_delete = true

  s3_buckets = {
    "bucket" = {
      name      = "my-app-txn-${random_id.suffix.hex}"
      folder_id = kvindo_folder.main.id
      tier      = "standard"
      region    = "ru-msk-1"
      quota_gib = 10
    }
  }

  s3_user_access_policies = {
    "rw" = {
      name        = "my-app-txn-rw"
      folder_id   = kvindo_folder.main.id
      policy_json = jsonencode({ ... })
    }
  }

  s3_users = {
    "app" = {
      id                = local.txn_user_app_id   # pre-generated ULID
      name              = "my-app-txn-user"
      folder_id         = kvindo_folder.main.id
      bucket_id         = local.txn_bucket_id
      access_policy_ids = [local.txn_policy_rw_id]
    }
  }
}

output "access_key" {
  value     = kvindo_transaction.main.s3_users["app"].info.access_key
  sensitive = true
}
```

## Example: VM on a Private Network

```hcl
resource "kvindo_vpc" "main" {
  name      = "app-net"
  folder_id = kvindo_folder.main.id
}

resource "kvindo_vpc_subnet" "main" {
  name      = "app-subnet"
  folder_id = kvindo_folder.main.id
  vpc_id    = kvindo_vpc.main.id
  ipv4_cidr = "10.10.0.0/24"
}

resource "kvindo_ssh_key" "main" {
  name       = "deploy-key"
  folder_id  = kvindo_folder.main.id
  public_key = file("~/.ssh/id_ed25519.pub")
}

resource "kvindo_vm" "web" {
  name          = "web-1"
  folder_id     = kvindo_folder.main.id
  vpc_subnet_id = kvindo_vpc_subnet.main.id
  ssh_key_ids   = [kvindo_ssh_key.main.id]
  offer_id      = var.vm_offer_id   # compute-offer ULID — see GET /api/v1/vm-offer in the swagger
  image_id      = var.vm_image_id   # OS-image ULID — see GET /api/v1/image
}

output "web_public_ip"  { value = kvindo_vm.web.info.public_ipv4 }
output "web_private_ip" { value = kvindo_vm.web.info.private_ipv4 }
```

## Example: Kubernetes Cluster

```hcl
resource "kvindo_kubernetes" "main" {
  name          = "app-cluster"
  folder_id     = kvindo_folder.main.id
  vpc_subnet_id = kvindo_vpc_subnet.main.id
  version       = "1.30"        # see swagger for supported versions
  tier          = "standard"
}

resource "kvindo_kubernetes_node_group" "workers" {
  name               = "workers"
  folder_id          = kvindo_folder.main.id
  kubernetes_id      = kvindo_kubernetes.main.id
  vpc_subnet_id      = kvindo_vpc_subnet.main.id
  desired_node_count = 3
  vm_offer_id        = var.vm_offer_id
  volume_offer_id    = var.volume_offer_id
  volume_size_gib    = 50
}

output "kube_api_server" { value = kvindo_kubernetes.main.info.api_server_url }
```

## Example: Managed PostgreSQL

```hcl
resource "kvindo_postgresql_standalone" "db" {
  name            = "app-db"
  folder_id       = kvindo_folder.main.id
  vpc_subnet_id   = kvindo_vpc_subnet.main.id
  version         = "16"          # see swagger for supported versions
  tier            = "standard"
  root_password   = var.db_root_password   # write-only — the API never returns this
  vm_offer_id     = var.vm_offer_id
  volume_offer_id = var.volume_offer_id
  volume_size_gib = 20
}

output "db_host" { value = kvindo_postgresql_standalone.db.info.private_ip_v4 }
output "db_port" { value = kvindo_postgresql_standalone.db.info.port }
```

## Resource Lifecycle

All resources go through states: `Scheduled → Reconciling → Reconciled`. The provider polls until `Reconciled` (30-minute timeout). If a poll times out while provisioning is in progress, import the resource instead of recreating it:

```bash
terraform import kvindo_s3_bucket.main <resource-id>
```

## Design Notes

A few decisions differ from a naive provider and are worth explaining:

### Field names mirror the API
See [Field Naming](#field-naming). Attribute names track the OpenAPI schema at
<https://cloud-api.kvindo.com/swagger> one-to-one (only casing changes). This keeps the
provider a thin, predictable mapping over the REST API rather than an opinionated
re-modeling — you can always cross-reference the swagger to find a field, its type, and
its accepted values.

### Read-only data lives in `info`
Everything the server computes — `state`, IP addresses, endpoints, generated
credentials — is grouped under one computed `info` object instead of being mixed into the
configurable top-level schema. This makes it obvious which attributes you set versus which
the API returns, and gives every resource the same access pattern: `<resource>.info.<field>`.

### `state` uses a terminal-gated plan modifier
`info.state` is volatile: the server moves a resource through
`scheduling → reconciling → stable` (or `schedulingfailed`). A plain `UseStateForUnknown`
would freeze a stale value into the plan and then fail apply with *"Provider produced
inconsistent result after apply"*; always recomputing it would show a perpetual
`(known after apply)` diff. Instead the provider freezes `state` **only when its prior value
is the terminal `stable`**, and leaves it `(known after apply)` while in flight. The result:
clean plans when settled, correct re-resolution while provisioning, and self-healing (≤1
apply) after an interrupted run.

### Optional vs. Optional+Computed
Fields the API echoes back are `Optional + Computed` (with `UseStateForUnknown`) so an unset
value adopts whatever the server assigns without churn. Fields the API never returns —
write-only secrets like `root_password`, or references that stay null until attached like
`floating_ip_id` — are plain `Optional`; marking them `Computed` would produce a permanent
`(known after apply)` diff because there is nothing to read back.

### IDs are ULIDs
Every resource ID is a lowercase 26-char Crockford-base32 ULID, generated client-side for
new resources. Client-side generation is what lets a `kvindo_transaction` reference a
sub-resource's ID before the API has created it.

### Transactions are two-phase
`kvindo_transaction` creates many sub-resources in one atomic API call. Sub-resources are
map attributes keyed by a name you choose; those keys are kept stable in state across
applies (matched back by ID), so editing one entry never churns the others.

### Resilient polling
Every create/update polls until the resource reaches a terminal state (30-minute timeout).
A `PUT` that races a still-reconciling resource (HTTP 422 `ResourceIsScheduling`, common
right after a Ctrl-C) is retried automatically once the resource settles, so an interrupted
apply recovers cleanly on the next run.

## Notes

- All IDs are ULIDs (lowercase 26-char Crockford base32), not UUIDs.
- S3 bucket names are globally unique — always use a `random_id` suffix.
- Transaction sub-resource map keys are stable across `terraform apply` — the key you choose in config (e.g. `"app"`) is preserved in state. 
