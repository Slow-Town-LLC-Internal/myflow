# Outputs file (outputs.tf)
output "keyring_name" {
  value = google_kms_key_ring.env_keyring.name
}

output "key_name" {
  value = google_kms_crypto_key.env_key.name
}

output "key_id" {
  value = google_kms_crypto_key.env_key.id
}


output "service_account_email" {
  value = google_service_account.kms_service_account.email
  description = "The email address of the service account created for KMS operations"
}
