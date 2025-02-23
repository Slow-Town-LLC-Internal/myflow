# GCP KMS Environment Manager

Securely manage environment variables using Google Cloud KMS and Cloud Storage.

## Prerequisites
- Google Cloud Project with enabled APIs:
  - Cloud KMS API
  - Cloud Storage API
- Terraform installed
- Go 1.21 or later
- GCP credentials configured (`gcloud auth application-default login`)

## Setup

1. Initialize and apply Terraform:
```bash
terraform init
terraform apply
terraform output > tf-output.log
```

2. Build the KMS manager:
```bash
go build -o kms-manager
# Optional: Install to $GOPATH/bin
go install
```

## Usage

Encrypt and upload `.env`:
```bash
./kms-manager -action encrypt
```

Download and decrypt to `.env`:
```bash
./kms-manager -action decrypt
```

### Notes
- Automatically creates timestamped backups of existing `.env` files
- Encrypted secrets are stored in GCS bucket `{project-id}-encrypted-secrets`
- Files are encrypted using GCP KMS keys

## Files
- `main.tf`: Terraform configuration for KMS and GCS setup
- `main.go`: Go program for encryption/decryption operations
- `tf-output.log`: Terraform output (required for kms-manager, git-ignored)

## Security
- All operations use Google Cloud KMS for encryption
- Encrypted data stored in versioned GCS bucket
- Local files maintained with 0600 permissions
- IAM-based access control
