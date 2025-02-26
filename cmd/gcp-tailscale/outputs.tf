# Output the public IP address (ephemeral)
output "public_ip" {
  value       = google_compute_instance.tailscale_node.network_interface.0.access_config.0.nat_ip
  description = "The ephemeral public IP address of the VM"
}

# Output Tailscale information
output "tailscale_hostname" {
  value       = var.tailscale_hostname
  description = "The Tailscale hostname for this exit node"
}



# Output instructions for next steps
output "next_steps" {
  value       = <<-EOT
    Setup completed:
    1. Your Tailscale exit node (${var.tailscale_hostname}) is configured and should appear in your Tailscale admin panel
    2. SSH to your instance if needed: ssh ${var.ssh_username}@${google_compute_instance.tailscale_node.network_interface.0.access_config.0.nat_ip}
    3. Check Tailscale status through Tailscale admin panel
    4. Verify exit node functionality by enabling exit node usage in your Tailscale admin console

    Note: This instance uses an ephemeral IP which may change if the VM is stopped/started.
    Use Tailscale to access it reliably once set up.
  EOT
  description = "Next steps after deployment"
}
