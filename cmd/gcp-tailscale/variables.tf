variable "tailscale_auth_key" {
  description = "Tailscale pre-authentication key"
  type        = string
  sensitive   = true
}

variable "tailscale_hostname" {
  description = "Hostname for this Tailscale node"
  type        = string
  default     = "gcp-exit-node"
}

variable "ssh_username" {
  description = "Username for SSH access"
  type        = string
  default     = "admin"
}

variable "ssh_public_key_path" {
  description = "Path to your SSH public key"
  type        = string
  default     = "~/.ssh/id_rsa.pub"
}

variable "allowed_ssh_cidr" {
  description = "IP CIDR range allowed for SSH access"
  type        = string
  default     = "0.0.0.0/0" # Warning: This allows SSH from anywhere, consider restricting to your IP
}

variable "gcp_project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "gcp_zone" {
  description = "GCP Zone for the instance"
  type        = string
  default     = "us-central1-a"
}
