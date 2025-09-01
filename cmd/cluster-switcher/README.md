# EKS Cluster Switcher

A command-line tool to switch between EKS clusters and local Kubernetes contexts with TouchID/Keychain integration support via aws-vault.

## Features

- 🔐 **TouchID Authentication**: Secure AWS credentials with macOS Keychain and TouchID
- 🔄 **Multiple Auth Methods**: Support for aws-vault, AWS SSO, or local contexts
- ⚡ **Credential Caching**: Configurable credential cache duration per environment
- 🌍 **Multi-Environment**: Support for EKS (prod/staging/dev) and local clusters (Orbstack/Docker)
- 🔒 **Security Levels**: Different security requirements per environment

## Installation

```bash
# Clone the repository
git clone https://github.com/Slow-Town-LLC/cluster-switcher.git
cd cluster-switcher

# Build the enhanced version
go build -o cluster-switcher cluster-switcher-enhanced.go

# Or install directly
go install github.com/Slow-Town-LLC/cluster-switcher@latest
```

### Prerequisites

```bash
# Install aws-vault for TouchID support (recommended)
brew install aws-vault

# Other requirements
# - AWS CLI v2
# - kubectl
# - Go 1.19 or later
```

## Configuration

Create a configuration file at `~/.config/eks-config.yaml`:

```yaml
# Global defaults
default_auth_method: aws-vault      # Options: aws-vault, sso
default_cache_duration: 15m         # Credential cache duration

environments:
  # Production - High security with TouchID
  prd:
    aws_profile: prd-admin
    eks_cluster: prd-use2-eks
    eks_region: us-east-2
    auth_method: aws-vault           # Use aws-vault with Keychain
    require_touchid: true            # Always require TouchID
    cache_duration: 5m               # Short cache for production

  # Staging - Moderate security
  stg:
    aws_profile: stg-admin
    eks_cluster: stg-use2-eks
    eks_region: us-east-2
    auth_method: aws-vault
    require_touchid: true
    cache_duration: 30m              # Longer cache for staging

  # Local Orbstack cluster - No AWS authentication
  local:
    is_local: true
    local_context: orbstack          # The kubectl context name
```

## Setting Up aws-vault

If you're migrating from AWS SSO to aws-vault for TouchID support:

### 1. Add your AWS credentials to aws-vault

```bash
# Add credentials for each profile (will be stored in Keychain)
aws-vault add prd-admin
aws-vault add stg-admin

# List configured profiles
aws-vault list
```

### 2. Test TouchID authentication

```bash
# This should trigger TouchID prompt
aws-vault exec prd-admin -- aws sts get-caller-identity
```

## Usage

### List all environments
```bash
cluster-switcher list
```

### Switch to environments
```bash
# Switch to production (triggers TouchID)
cluster-switcher prd

# Switch to staging (triggers TouchID)
cluster-switcher stg

# Switch to local Orbstack (no TouchID needed)
cluster-switcher local
```

### After switching
The tool will output export commands for your shell:
```bash
export AWS_PROFILE=prd-admin
export KUBECONFIG=~/.kube/configs/eks-prd/config
```

## How It Works

### With aws-vault (TouchID)
1. Credentials stored securely in macOS Keychain
2. TouchID required to access credentials
3. Temporary AWS credentials generated (cached for configured duration)
4. kubectl commands automatically trigger TouchID when cache expires

### With SSO (Legacy)
1. Browser-based authentication
2. Session cached by AWS CLI
3. No TouchID integration

### With Local Contexts
1. Direct context switching
2. No AWS authentication needed
3. Works with Orbstack, Docker Desktop, etc.

## Security Benefits

- **No plaintext credentials**: All AWS credentials stored in Keychain
- **Hardware authentication**: TouchID ensures physical presence
- **Time-limited access**: Credentials expire after cache duration
- **Per-environment policies**: Production can require stricter authentication

## Troubleshooting

### TouchID not working
```bash
# Ensure aws-vault is using the correct backend
aws-vault --backend=keychain list
```

### Profile not found in aws-vault
```bash
# Add the profile
aws-vault add <profile-name>
```

### Clear cached credentials
```bash
# Remove specific profile session
aws-vault clear <profile-name>

# Remove all sessions
aws-vault clear --sessions
```

## Environment Variables

- `EKS_CONFIG`: Custom config file path (default: `~/.config/eks-config.yaml`)
- `AWS_VAULT_BACKEND`: Force specific backend (default: `keychain` on macOS)
- `AWS_VAULT_PROMPT`: TouchID prompt method (default: `osascript`)

## Development

```bash
# Run tests
go test ./...

# Build
go build -o cluster-switcher cluster-switcher-enhanced.go

# Install locally
go install .
```

## License

MIT License
