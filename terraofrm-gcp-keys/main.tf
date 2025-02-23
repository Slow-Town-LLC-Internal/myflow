# main.tf

terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 4.0"
    }
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}


# Create GCS bucket for encrypted secrets
resource "google_storage_bucket" "secrets_bucket" {
  name          = "${var.project_id}-encrypted-secrets"
  location      = var.region
  force_destroy = true

  uniform_bucket_level_access = true

  versioning {
    enabled = true
  }
}


# Create dedicated service account for KMS operations
resource "google_service_account" "kms_service_account" {
  account_id   = "kms-manager-sa"
  display_name = "KMS Manager Service Account"
  description  = "Service account for managing KMS operations"
}


# Create a KMS keyring
resource "google_kms_key_ring" "env_keyring" {
  name     = "env-secrets-keyring"
  location = var.region
}

# Create a KMS key
resource "google_kms_crypto_key" "env_key" {
  name     = "env-secrets-key"
  key_ring = google_kms_key_ring.env_keyring.id

  # Optional: Configure rotation period
  rotation_period = "7776000s"  # 90 days

  # Optional: Configure purpose
  purpose = "ENCRYPT_DECRYPT"

  # Optional: Configure version template
  version_template {
    algorithm = "GOOGLE_SYMMETRIC_ENCRYPTION"
    protection_level = "SOFTWARE"
  }
}

# IAM binding for the service account to use the key
resource "google_kms_crypto_key_iam_binding" "crypto_key" {
  crypto_key_id = google_kms_crypto_key.env_key.id
  role          = "roles/cloudkms.cryptoKeyEncrypterDecrypter"

  members = [
    "serviceAccount:${google_service_account.kms_service_account.email}",
  ]
}

# Grant additional required permissions to the service account
resource "google_project_iam_member" "kms_sa_permissions" {
  project = var.project_id
  role    = "roles/cloudkms.viewer"
  member  = "serviceAccount:${google_service_account.kms_service_account.email}"
}



# Grant additional required permissions to the service account
resource "google_storage_bucket_iam_member" "secrets_bucket_access" {
  bucket = google_storage_bucket.secrets_bucket.name
  role   = "roles/storage.objectViewer"
  member = "serviceAccount:${google_service_account.kms_service_account.email}"
}

resource "google_storage_bucket_iam_member" "secrets_bucket_writer" {
  bucket = google_storage_bucket.secrets_bucket.name
  role   = "roles/storage.objectCreator"
  member = "serviceAccount:${google_service_account.kms_service_account.email}"
}
