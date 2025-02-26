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
  zone    = var.gcp_zone
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
  source_ranges = [var.allowed_ssh_cidr] # Configurable via variables.tf
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
  machine_type = "e2-micro" # Cheapest option sufficient for Tailscale exit node
  tags         = ["tailscale", "no-public-ports"]
  project      = var.gcp_project_id
  zone         = var.gcp_zone

  boot_disk {
    initialize_params {
      image = "debian-cloud/debian-11" # Using Debian 11, adjust as needed
      size  = 10                       # Reduced to 10GB disk size
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
    enable-oslogin         = "FALSE"
    block-project-ssh-keys = "TRUE" # Only use instance-specific SSH keys
    ssh-keys               = "${var.ssh_username}:${file(var.ssh_public_key_path)}"
  }

  # Install Tailscale using startup script with templatefile function
  metadata_startup_script = templatefile("init-ts-host.tftpl", {
    tailscale_auth_key = var.tailscale_auth_key,
    tailscale_hostname = var.tailscale_hostname
  })

  # Shielded VM settings for enhanced security
  shielded_instance_config {
    enable_secure_boot          = true
    enable_vtpm                 = true
    enable_integrity_monitoring = true
  }

  service_account {
    scopes = ["compute-ro", "storage-ro"] # Minimal required scopes
  }
}
