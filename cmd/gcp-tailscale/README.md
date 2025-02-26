# Tailscale Exit Node on GCP

This Terraform configuration creates a secure GCP instance configured as a Tailscale exit node.

## Prerequisites

- [Terraform](https://www.terraform.io/downloads.html) installed
- [Google Cloud SDK](https://cloud.google.com/sdk/docs/install) installed and authenticated
- A [Tailscale](https://tailscale.com) account with admin access
- A Tailscale auth key (preferably a pre-auth key)

## Setup

1. Clone this repository
   ```
   git clone <repo-url>
   cd tailscale-exit-node
   ```

2. Generate a Tailscale pre-auth key from your [Tailscale Admin Console](https://login.tailscale.com/admin/settings/keys)
   - Set it as an environment variable: `export TS_API_KEY=tskey-your-key-here`

3. Initialize Terraform
   ```
   terraform init
   ```

4. Apply the configuration
   ```
   terraform apply \
     -var="tailscale_auth_key=${TS_API_KEY}" \
     -var="gcp_project_id=your-gcp-project-id"
   ```

## Configuration Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `gcp_project_id` | GCP Project ID | - |
| `gcp_zone` | GCP Zone for the instance | `us-central1-a` |
| `tailscale_auth_key` | Tailscale pre-auth key | - |
| `tailscale_hostname` | Hostname for this Tailscale node | `gcp-exit-node` |
| `ssh_username` | Username for SSH access | `admin` |
| `ssh_public_key_path` | Path to your SSH public key | `~/.ssh/id_rsa.pub` |
| `allowed_ssh_cidr` | IP CIDR range allowed for SSH access | `0.0.0.0/0` |

## Security Features

- SSH password authentication disabled
- Firewall allows only SSH and ICMP
- Automatic security updates
- Shielded VM with secure boot
- IP forwarding enabled for exit node functionality
- Minimal VM size (e2-micro) for cost efficiency

## Teardown

To destroy the infrastructure:
```
terraform destroy \
  -var="tailscale_auth_key=${TS_API_KEY}" \
  -var="gcp_project_id=your-gcp-project-id"
```

## Note

The VM uses an ephemeral public IP to reduce costs. After initial setup, you should access the VM through Tailscale rather than SSH over the public internet.
