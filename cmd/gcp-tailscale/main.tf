terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 4.0"
    }
  }
}

# Configure the Google Cloud provider - using existing gcloud authentication
provider "google" {
  # Empty configuration will use your existing gcloud CLI authentication
  # But we'll specify the project explicitly
  project = var.gcp_project_id
  zone = var.gcp_zone
}

# Create VPC network with custom firewall rules
resource "google_compute_network" "tailscale_network" {
  name                    = "tailscale-network"
  auto_create_subnetworks = true
  project                 = var.gcp_project_id
}

# Firewall rule to allow only SSH initially for setup
resource "google_compute_firewall" "allow_ssh" {
  name    = "allow-ssh"
  network = google_compute_network.tailscale_network.name
  project = var.gcp_project_id

  allow {
    protocol = "tcp"
    ports    = ["22"]
  }

  # Restrict SSH to specific IP addresses
  source_ranges = [var.allowed_ssh_cidr]  # Configurable via variables.tf
}

# Firewall rule to allow ICMP (ping) for basic connectivity testing
resource "google_compute_firewall" "allow_icmp" {
  name    = "allow-icmp"
  network = google_compute_network.tailscale_network.name
  project = var.gcp_project_id

  allow {
    protocol = "icmp"
  }

  # Allow ICMP from anywhere (can be restricted if needed)
  source_ranges = ["0.0.0.0/0"]
}

# Create the VM instance
resource "google_compute_instance" "tailscale_node" {
  name         = "tailscale-node"
  machine_type = "e2-micro"  # Cheapest option sufficient for Tailscale exit node
  tags         = ["tailscale", "no-public-ports"]
  project      = var.gcp_project_id
  zone         = var.gcp_zone

  boot_disk {
    initialize_params {
      image = "debian-cloud/debian-11"  # Using Debian 11, adjust as needed
      size  = 10  # Reduced to 10GB disk size
    }
  }

  network_interface {
    network = google_compute_network.tailscale_network.name
    access_config {
      # Using ephemeral IP instead of static IP
    }
  }

  # Ensure OS Login is disabled to use SSH keys in metadata
  metadata = {
    enable-oslogin = "FALSE"
    block-project-ssh-keys = "TRUE"  # Only use instance-specific SSH keys
    ssh-keys = "${var.ssh_username}:${file(var.ssh_public_key_path)}"
  }

  # Install Tailscale using startup script - improved version
  metadata_startup_script = <<-EOT
    #!/bin/bash

    # Set variables using environment variables
    tailscale_auth_key="${var.tailscale_auth_key}"
    tailscale_hostname="${var.tailscale_hostname}"

    # Create a log file
    exec > >(tee /var/log/tailscale-setup.log) 2>&1

    echo "Starting Tailscale setup script at $$(date)"

    # Update system
    echo "Updating system packages..."
    apt-get update

    # Install required packages
    echo "Installing required packages..."
    apt-get install -y curl iptables-persistent apt-transport-https gnupg

    # Secure SSH configuration
    echo "Configuring SSH..."
    sed -i 's/PermitRootLogin yes/PermitRootLogin no/' /etc/ssh/sshd_config
    sed -i 's/#PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config
    sed -i 's/#PubkeyAuthentication yes/PubkeyAuthentication yes/' /etc/ssh/sshd_config
    systemctl restart sshd

    # Configure firewall (iptables)
    echo "Configuring firewall..."
    # Flush existing rules
    iptables -F

    # Set default policies
    iptables -P INPUT DROP
    iptables -P FORWARD DROP
    iptables -P OUTPUT ACCEPT

    # Allow established connections
    iptables -A INPUT -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT

    # Allow local loopback
    iptables -A INPUT -i lo -j ACCEPT

    # Allow SSH (temporarily for setup)
    iptables -A INPUT -p tcp --dport 22 -j ACCEPT

    # Allow ICMP (ping)
    iptables -A INPUT -p icmp -j ACCEPT

    # Save iptables rules
    netfilter-persistent save

    # Add Tailscale repository
    echo "Adding Tailscale repository..."
    curl -fsSL https://pkgs.tailscale.com/stable/debian/bullseye.noarmor.gpg | tee /usr/share/keyrings/tailscale-archive-keyring.gpg >/dev/null
    curl -fsSL https://pkgs.tailscale.com/stable/debian/bullseye.tailscale-keyring.list | tee /etc/apt/sources.list.d/tailscale.list

    # Update and install Tailscale
    echo "Installing Tailscale..."
    apt-get update
    apt-get install -y tailscale

    # Enable IP forwarding for exit node functionality
    echo "Enabling IP forwarding..."
    echo 'net.ipv4.ip_forward = 1' | tee -a /etc/sysctl.conf
    echo 'net.ipv6.conf.all.forwarding = 1' | tee -a /etc/sysctl.conf
    sysctl -p /etc/sysctl.conf

    # Set up Tailscale as an exit node using provided pre-auth key
    echo "Configuring Tailscale as exit node..."
    tailscale up --authkey="$$tailscale_auth_key" --hostname="$$tailscale_hostname" --advertise-exit-node

    # Configure Tailscale to start on boot
    echo "Enabling Tailscale service..."
    systemctl enable tailscaled

    # Set up automatic updates for security
    echo "Configuring automatic updates..."
    apt-get install -y unattended-upgrades
    cat > /etc/apt/apt.conf.d/20auto-upgrades <<EOF
APT::Periodic::Update-Package-Lists "1";
APT::Periodic::Unattended-Upgrade "1";
APT::Periodic::AutocleanInterval "7";
EOF

    # Configure unattended-upgrades
    cat > /etc/apt/apt.conf.d/50unattended-upgrades <<EOF
Unattended-Upgrade::Allowed-Origins {
  "$${distro_id}:$${distro_codename}";
  "$${distro_id}:$${distro_codename}-security";
  "$${distro_id}ESMApps:$${distro_codename}-apps-security";
  "$${distro_id}ESM:$${distro_codename}-infra-security";
  "TailscaleOfficial:stable";
};
Unattended-Upgrade::Package-Blacklist {
};
Unattended-Upgrade::Automatic-Reboot "true";
Unattended-Upgrade::Automatic-Reboot-Time "02:00";
EOF

    # Enable automatic updates
    systemctl enable unattended-upgrades
    systemctl start unattended-upgrades

    # Verify Tailscale status
    echo "Tailscale setup complete. Current status:"
    tailscale status

    echo "Setup completed at $$(date)"
  EOT


  # Shielded VM settings for enhanced security
  shielded_instance_config {
    enable_secure_boot          = true
    enable_vtpm                 = true
    enable_integrity_monitoring = true
  }

  service_account {
    scopes = ["compute-ro", "storage-ro"]  # Minimal required scopes
  }
}
