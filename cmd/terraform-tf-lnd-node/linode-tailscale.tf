terraform {
  required_providers {
    linode = {
      source  = "linode/linode"
      version = "~> 2.5"
    }
  }
}

provider "linode" {
  token = var.linode_token
}

variable "linode_token" {
  description = "Linode API token"
  type        = string
  default     = ""
}

variable "tailscale_authkey" {
  description = "Tailscale auth key for automatic connection"
  type        = string
  sensitive   = true
}

variable "ssh_public_key" {
  description = "SSH public key for access"
  type        = string
}

variable "advertise_routes" {
  description = "Additional routes to advertise (comma-separated, e.g., '192.168.1.0/24,10.0.0.0/8')"
  type        = string
  default     = ""
}

# Get the latest Ubuntu LTS image
data "linode_images" "ubuntu" {
  filter {
    name   = "label"
    values = ["Ubuntu 24.04 LTS"]
  }
  filter {
    name   = "is_public"
    values = ["true"]
  }
}

resource "linode_instance" "tailscale_node" {
  label       = "tailscale-micro"
  region      = "us-east"  # Change to your preferred region
  type        = "g6-nanode-1"  # Cheapest instance type
  image       = data.linode_images.ubuntu.images[0].id
  
  # Set root password (required but we'll use SSH keys)
  root_pass = random_password.root_password.result
  
  # Add SSH key
  authorized_keys = [var.ssh_public_key]
  
  # Enable backups if needed (adds cost)
  backups_enabled = false
  
  # Boot config
  booted = true
  
  # Add startup script to install and configure Tailscale
  metadata {
    user_data = base64encode(<<-EOT
#cloud-config
hostname: pezware-inc
users:
  - name: pezware
    groups: sudo
    shell: /bin/bash
    sudo: ['ALL=(ALL) NOPASSWD:ALL']
    ssh_authorized_keys:
      - ${var.ssh_public_key}

package_update: true
package_upgrade: true

packages:
  - curl
  - wget
  - htop
  - vim
  - ufw

runcmd:
  # Set hostname
  - hostnamectl set-hostname pezware-inc
  
  # Install Tailscale
  - curl -fsSL https://tailscale.com/install.sh | sh
  
  # Enable IP forwarding for exit node functionality
  - echo 'net.ipv4.ip_forward = 1' | tee -a /etc/sysctl.conf
  - echo 'net.ipv6.conf.all.forwarding = 1' | tee -a /etc/sysctl.conf
  - sysctl -p
  
  # Start Tailscale with auth key, SSH, and exit node advertisement
  - tailscale up --authkey=${var.tailscale_authkey} --ssh --advertise-exit-node --accept-routes --hostname=pezware-inc
  
  # Configure firewall - Block SSH from public, allow only via Tailscale
  - ufw default deny incoming
  - ufw default allow outgoing
  - ufw allow 41641/udp  # Tailscale
  - ufw allow in on tailscale0  # Allow all traffic on Tailscale interface
  - ufw --force enable
  
  # Disable SSH from public interface (only allow via Tailscale)
  - sed -i 's/#ListenAddress 0.0.0.0/ListenAddress 100.64.0.0/g' /etc/ssh/sshd_config
  - echo "ListenAddress $(tailscale ip -4)" >> /etc/ssh/sshd_config
  - systemctl restart sshd
  
  # Ensure Tailscale starts on boot
  - systemctl enable tailscaled
  - systemctl start tailscaled

# Reboot after kernel updates
power_state:
  mode: reboot
  condition: True
EOT
    )
  }
  
  tags = ["tailscale", "micro"]
}

resource "random_password" "root_password" {
  length  = 32
  special = true
}

output "instance_ip" {
  value = linode_instance.tailscale_node.ip_address
}

output "instance_id" {
  value = linode_instance.tailscale_node.id
}

output "instance_status" {
  value = linode_instance.tailscale_node.status
}

output "root_password" {
  value     = random_password.root_password.result
  sensitive = true
}

output "setup_instructions" {
  value = <<-EOT
  
  Tailscale Setup Complete!
  
  1. Your instance public IP: ${linode_instance.tailscale_node.ip_address}
     ⚠️  SSH is DISABLED on public IP for security!
  
  2. SSH access (Tailscale ONLY):
     - Via Tailscale: ssh pezware@pezware-inc
     - Or use: tailscale ssh pezware-inc
  
  3. Exit Node Configuration:
     - The node is advertising as an exit node
     - You must approve it in the Tailscale admin console:
       https://login.tailscale.com/admin/machines
     - Look for "pezware-inc" and click "Edit route settings"
     - Toggle "Use as exit node" to approve
  
  4. To use this as your exit node from other devices:
     - Run: tailscale up --exit-node=pezware-inc
     - Or select it in the Tailscale app
  
  5. Check Tailscale status:
     - Connect via: tailscale ssh pezware-inc
     - Run: sudo tailscale status
     - Run: sudo tailscale ip -4  # Get Tailscale IP
  
  6. The instance will reboot once after initial setup to apply kernel updates.
  
  7. Your Tailscale network (100.x.x.x) is fully compatible with this setup.
  
  EOT
}