# Developer Workstation Security Guide

## Table of Contents
1. [Baseline Security Requirements](#baseline-security-requirements)
2. [Disk Encryption](#1-disk-encryption)
3. [Secrets Management](#2-secrets-management)
4. [Human Review Gates](#3-human-review-gates)
5. [Network Security](#4-network-security-optional)
6. [Backup and Key Rotation](#5-backup-and-key-rotation)
7. [Browser Security](#6-browser-security)
8. [Development Containers Security](#7-development-containers-security)
9. [Supply Chain Hygiene](#8-supply-chain-hygiene)
10. [Resources](#resources)

## Baseline Security Requirements

1. **Encrypted disk** - Full‑disk encryption enabled (FileVault/LUKS/BitLocker) or ephemeral live environment
2. **No secrets in files** - Use OS keychain or secrets manager
3. **Human review gate** - Hardware token or biometric approval
4. **Network monitoring** - Optional egress filtering/monitoring; use TLS interception (MITM) only for debugging with care
5. **Backup and rekey** - Regular rotation with secure backup strategy
6. **Access control** - Screen lock after 5 minutes inactivity, password/biometric unlock
7. **No production data locally** - Production data/logs must be deleted immediately after debugging


---

## Security Controls Summary Table

| Auth Type | Storage Method | Human Review | Rotation Schedule | Backup Strategy | Additional Controls |
|-----------|----------------|--------------|-------------------|-----------------|---------------------|
| **AWS/Azure OIDC** | OS Keychain (temp tokens) | MFA at auth | Auto-refresh on reauth | N/A (ephemeral) | Session timeout, assume role patterns |
| **System Passwords** | OS Keychain | Touch ID/YubiKey | Quarterly or per org policy | Password manager sync | Complexity requirements, history check |
| **YubiKey** | Hardware secure element | PIN + physical touch | Change PIN quarterly | 2+1 backup keys | Custom PUK, attestation certificates |
| **API Keys (Personal)** | OS Keychain | Touch ID/biometric approval | Every 60 days | Encrypted backup | Scope limitations, usage monitoring |
| **System API Keys** | AWS Secrets Manager/Azure KV | Approval workflow | 30-90 days or vendor requirement | Key versioning | Audit logging, least privilege |
| **SSH Keys** | ssh-agent + keychain | Passphrase + Touch ID | Annual | Encrypted backup | Ed25519 or P‑256 (hardware) or RSA‑2048+, hardware key preferred |
| **Git Credentials** | Git credential helper | 2FA/SSO | Follow OIDC/PAT policy | Via identity provider | Sign commits, SSH over HTTPS |
| **Database Credentials** | HashiCorp Vault/Keychain | MFA for access | Dynamic/30 days | Via secrets manager | Connection pooling, SSL/TLS required |
| **OAuth Tokens** | OS Keychain | App-specific consent | Per token expiry | Via identity provider | Scope minimization, refresh token rotation |
| **TLS/SSL Client Certs** | OS Keychain/Hardware token | PIN/passphrase | Annual or before expiry | Encrypted backup, HSM | Certificate pinning, OCSP validation |

## Notes:
- **Environment Variables**: Should never store secrets directly. Load from keychain/vault at runtime into memory only
- **Keys**: Prefer Ed25519, P‑256 (ECC), or RSA‑2048+ for asymmetric keys
- **Screen Lock**: Auto-lock after 5 minutes of inactivity, require password/biometric to unlock
---

## 1. Disk Encryption

### macOS - FileVault
```bash
# Enable FileVault
sudo fdesetup enable

# Verify FileVault status
sudo fdesetup status

# CRITICAL: Store recovery key securely - NEVER on the same machine!
# Option 1: Personal recovery key (write down and store in secure location)
# Option 2: Institutional recovery key (for enterprise environments)
sudo fdesetup validaterecovery

# Check System Integrity Protection (SIP) status (Apple Silicon/T2 Macs)
csrutil status

# Note: Secure Boot settings are managed via Startup Security Utility (no direct CLI)

# Check encryption cipher (should be AES-XTS-128 or AES-XTS-256)
diskutil apfs list | grep "Encryption Type"
```

### Linux - LUKS
```bash
# Check current encryption status
lsblk -o NAME,FSTYPE,SIZE,MOUNTPOINT,UUID | grep -E "crypto|LUKS"

# For NEW installation: Use installer's full disk encryption (recommended)

# For EXISTING system - verify LUKS version (LUKS2 recommended)
sudo cryptsetup luksDump /dev/sdXY | grep Version  # Replace sdXY with your partition

# Option 1: Full disk encryption (requires reinstall)
# Use distribution installer with encryption option

# Option 2: Encrypt home directory only (less secure)
sudo apt-get install cryptsetup
# Create encrypted container for home
sudo dd if=/dev/zero of=/home/.encrypted bs=1M count=10240  # 10GB, adjust size
sudo cryptsetup luksFormat /home/.encrypted
sudo cryptsetup open /home/.encrypted home_crypt
sudo mkfs.ext4 /dev/mapper/home_crypt

# Note: ecryptfs is deprecated - use LUKS containers or systemd-homed
```

### Alternative: Linux Live USB (High-Security Environments)
```bash
# Create persistent encrypted live USB for sensitive work
# WARNING: This is for temporary high-security sessions, not daily use

# Step 1: Write ISO to USB (replace sdX with your USB device)
sudo dd if=ubuntu-24.04-live.iso of=/dev/sdX bs=4M status=progress conv=fsync

# Step 2: Create encrypted persistence partition
sudo fdisk /dev/sdX  # Create new partition for persistence
sudo cryptsetup luksFormat --type luks2 /dev/sdX2
sudo cryptsetup open /dev/sdX2 live-persistence

# Step 3: Format and label for persistence
sudo mkfs.ext4 /dev/mapper/live-persistence
sudo e2label /dev/mapper/live-persistence persistence

# Step 4: Configure persistence (boot USB first, then run)
sudo mkdir -p /mnt/persistence
sudo mount /dev/mapper/live-persistence /mnt/persistence
echo "/ union" | sudo tee /mnt/persistence/persistence.conf

# Note: All changes stored encrypted, wiped on USB removal
```

### Windows - BitLocker (for WSL2 users)
```powershell
# Enable BitLocker (Windows Pro/Enterprise)
manage-bde -on C: -encryptionmethod XTS-AES-256 -recoverypassword

# Verify BitLocker status
manage-bde -status

# Store recovery key in Microsoft account or AD (enterprise)
# NEVER store recovery key on the same machine
```

---

## 2. Secrets Management

### macOS Keychain Integration

```bash
# Store API key (will prompt for Touch ID on retrieval)
security add-internet-password \
  -a "api-user" \
  -s "api.company.com" \
  -w "API_KEY" \
  -r htps \
  -T /usr/bin/curl \
  -U  # Update if exists

# Retrieve API key (requires Touch ID/password)
security find-internet-password -s "api.company.com" -a "api-user" -w

# Store generic password
security add-generic-password \
  -a "$USER" \
  -s "MyApp" \
  -w "SECRET_VALUE" \
  -U

# Retrieve for use in scripts
export API_KEY=$(security find-generic-password -a "$USER" -s "MyApp" -w)

# Delete a secret
security delete-generic-password -a "$USER" -s "MyApp"

# List all items (metadata only)
security dump-keychain | grep "keychain:" -A 5
```

### Linux Secret Service (GNOME Keyring/KWallet)

```bash
# GNOME Keyring (Ubuntu/Debian/Fedora)
# Install secret-tool
sudo apt-get install libsecret-tools gnome-keyring

# Store secret (will auto-unlock with login password)
echo -n "MY_SECRET" | secret-tool store \
  --label="My Credentials" \
  service vendor_api \
  username "$USER"

# Retrieve secret
secret-tool lookup service vendor_api username "$USER"

# Delete secret
secret-tool clear service vendor_api username "$USER"

# Use with environment variables
export VENDOR_API_KEY=$(secret-tool lookup service vendor_api username "$USER")

# GUI tool for management
seahorse &  # GNOME Passwords and Keys

# KDE KWallet (KDE Plasma)
# Use KWallet Manager GUI (kwalletmanager) to store secrets; CLI flows vary by distro
sudo apt-get install kwalletmanager
# Note: Configure auto-unlock in System Settings > KDE Wallet
```

### Cloud Provider OIDC Integration

#### AWS SSO/OIDC Configuration
```bash
# Install aws-vault for secure credential management
brew install aws-vault  # macOS
# Linux: Download from GitHub releases

# Configure AWS SSO profile in ~/.aws/config
cat >> ~/.aws/config << 'EOF'
[profile development]
sso_start_url = https://company.awsapps.com/start
sso_region = us-east-1
sso_account_id = 123456789012
sso_role_name = DeveloperAccess
region = us-east-1
EOF

# Login with aws-vault (triggers browser for SSO)
aws-vault login development

# Execute commands with temporary credentials
aws-vault exec development -- aws s3 ls

# Start credential server for GUI apps (optional)
aws-vault exec --server development

# Configure keychain backend (macOS)
export AWS_VAULT_BACKEND=keychain
export AWS_VAULT_KEYCHAIN_NAME=aws-vault

# Set session durations
export AWS_SESSION_TTL=4h
export AWS_ASSUME_ROLE_TTL=1h

# For MFA with aws-vault
export AWS_VAULT_PROMPT=osascript  # macOS Touch ID/dialog

# Chain roles for cross-account access
cat >> ~/.aws/config << 'EOF'
[profile production]
source_profile = development
role_arn = arn:aws:iam::987654321098:role/ProductionAccess
mfa_serial = arn:aws:iam::123456789012:mfa/username
EOF

# Use production profile with MFA
aws-vault exec production -- aws s3 ls
```

#### Azure CLI with OIDC
```bash
# Install Azure CLI
curl -sL https://aka.ms/InstallAzureCLIDeb | sudo bash

# Login with browser-based OIDC
az login

# Login with device code (for headless/SSH sessions)
az login --use-device-code

# List subscriptions and set default
az account list --output table
az account set --subscription "SUBSCRIPTION_ID"

# WARNING: Avoid service principals on dev machines when possible
# If required, store securely in keychain, never in files
```

#### GCP/Google Cloud Authentication
```bash
# Install gcloud CLI
curl https://sdk.cloud.google.com | bash

# Login with browser-based OAuth
gcloud auth login

# Application default credentials (for local development)
gcloud auth application-default login

# Set default project
gcloud config set project PROJECT_ID

# Use impersonation instead of service account keys
gcloud config set auth/impersonate_service_account SA_EMAIL@PROJECT.iam.gserviceaccount.com

# WARNING: Never store service account keys in files
# If absolutely necessary, use keychain instead
```

### Docker Credentials Management

```bash
# NEVER store credentials in ~/.docker/config.json in plain text

# Configure Docker to use OS keychain
cat > ~/.docker/config.json << 'EOF'
{
  "credsStore": "osxkeychain"  # macOS
  # "credsStore": "secretservice"  # Linux GNOME
  # "credsStore": "kwallet"  # Linux KDE
}
EOF

# Login (credentials automatically stored in keychain)
docker login registry.company.com

# For CI/CD, use short-lived tokens via stdin
echo $CI_REGISTRY_TOKEN | docker login -u $CI_REGISTRY_USER --password-stdin registry.company.com
```

---

## 3. Human Review Gates

Human review gates require physical presence and conscious approval for sensitive operations, preventing automated credential theft.

### macOS SSH Key Management with Secretive
```bash
# Install Secretive for Secure Enclave SSH key storage (Apple Silicon/T2 Macs only)
brew install --cask secretive

# Launch Secretive
open -a Secretive

# Generate new SSH key in Secure Enclave (via Secretive GUI)
# Keys are hardware-bound and require Touch ID for every use

# Configure SSH to use Secretive's agent
cat >> ~/.ssh/config << 'EOF'
Host *
    IdentityAgent "~/Library/Containers/com.maxgoedjen.Secretive.SecretAgent/Data/socket.ssh"
EOF

# Export public key from Secretive
# Copy from Secretive app or use:
ssh-add -L

# Note: For Intel Macs, use ssh-agent with passphrase-protected keys
```

### macOS Keychain with Touch ID (using keymaster)
```bash
# Install keymaster for Touch ID keychain access
brew tap johnthethird/keymaster
brew install keymaster

# Store secrets with Touch ID protection
keymaster set AWS_SECRET_KEY
# Enter value when prompted, will require Touch ID on retrieval

# Use in scripts - prompts for Touch ID
export AWS_SECRET_ACCESS_KEY=$(keymaster get AWS_SECRET_KEY)

# For standard keychain with Touch ID requirement
# First, add to keychain normally
security add-generic-password \
  -a "$USER" \
  -s "AWS" \
  -w "YOUR_SECRET_KEY"

# Then always allow keymaster binary when prompted
# Future access will require Touch ID
```

### macOS Touch ID for sudo
```bash
# Enable Touch ID for sudo
sudo sed -i '' '1s/^/auth       sufficient     pam_tid.so\n/' /etc/pam.d/sudo

# Test
sudo -k && sudo echo "Touch ID working"

# Note: macOS updates reset this. Add to ~/.zshrc to check on startup:
if ! grep -q "pam_tid.so" /etc/pam.d/sudo; then
  echo "⚠️  Touch ID for sudo disabled. Run: sudo sed -i '' '1s/^/auth       sufficient     pam_tid.so\\n/' /etc/pam.d/sudo"
fi
```

### YubiKey Setup

```bash
# Change default PIN and PUK (Management Key)
ykman piv access change-pin -P 123456 -n <NEW_PIN>
ykman piv access change-puk -p 12345678 -n <NEW_PUK>
ykman piv access change-management-key \
  -m 010203040506070801020304050607080102030405060708 \
  -n <NEW_MGMT_KEY>

# Generate PIV keys with touch policy (slots: 9a=auth, 9c=signing, 9d=encryption)
ykman piv keys generate -a RSA2048 -F PEM \
  --touch-policy ALWAYS --pin-policy ONCE 9a public-auth.pem
ykman piv keys generate -a RSA2048 -F PEM \
  --touch-policy ALWAYS --pin-policy ONCE 9c public-sign.pem
ykman piv keys generate -a RSA2048 -F PEM \
  --touch-policy ALWAYS --pin-policy ONCE 9d public-enc.pem

# Generate self-signed certificates
ykman piv certificates generate -s 9a -S "/CN=YubiKey Auth" public-auth.pem
ykman piv certificates generate -s 9c -S "/CN=YubiKey Sign" public-sign.pem
ykman piv certificates generate -s 9d -S "/CN=YubiKey Encrypt" public-enc.pem

# Setup backup YubiKeys (2+1 strategy: Primary + Secondary + Emergency)
# Export certificates from primary
for slot in 9a 9c 9d; do
  ykman piv certificates export $slot cert-$slot.pem
done
# Import to backup YubiKey (requires generating same keys on backup)
# ykman piv certificates import 9a cert-9a.pem  # Run on backup YubiKey
```

### Linux PAM with YubiKey
```bash
# Install YubiKey PAM module
sudo apt-get install libpam-yubico

# Configure PAM
echo "auth required pam_yubico.so mode=challenge-response" | \
  sudo tee -a /etc/pam.d/sudo

# Setup challenge-response
ykpersonalize -2 -ochal-resp -ochal-hmac -ohmac-lt64 -oserial-api-visible
```

---

## 4. Network Security (Optional)

### mitmproxy Configuration

```bash
# Install mitmproxy
brew install mitmproxy  # macOS
sudo apt-get install mitmproxy  # Linux

# Create log directory
mkdir -p ~/security/logs

# Create whitelist script
cat > ~/security/mitmproxy-whitelist.py << 'EOF'
from mitmproxy import http
import json

WHITELIST = [
    "*.company.com",
    "api.github.com",
    "*.amazonaws.com",
    "*.azure.com",
    "registry.npmjs.org",  # npm
    "pypi.org",  # Python packages
    "*.docker.io"  # Docker Hub
]

PAYLOAD_LIMIT = 1024 * 100  # 100KB

def request(flow: http.HTTPFlow) -> None:
    host = flow.request.pretty_host

    # Check whitelist
    if not any(match_domain(host, pattern) for pattern in WHITELIST):
        flow.response = http.Response.make(
            403,
            b"Domain not whitelisted",
            {"Content-Type": "text/plain"}
        )

    # Check payload size
    if len(flow.request.content or b'') > PAYLOAD_LIMIT:
        print(f"⚠️  Large payload detected: {len(flow.request.content)} bytes to {host}")
        # Log for review
        with open("~/security/logs/large-payloads.log", "a") as f:
            f.write(f"{flow.request.timestamp_start}: {host} - {len(flow.request.content)} bytes\n")

def match_domain(domain, pattern):
    if pattern.startswith("*."):
        return domain.endswith(pattern[2:]) or domain == pattern[2:]
    return domain == pattern
EOF

# Run proxy
mitmproxy -s ~/security/mitmproxy-whitelist.py --set confdir=~/.mitmproxy

# Configure system proxy
# macOS
networksetup -setwebproxy "Wi-Fi" 127.0.0.1 8080
networksetup -setsecurewebproxy "Wi-Fi" 127.0.0.1 8080

# Linux
export http_proxy=http://127.0.0.1:8080
export https_proxy=http://127.0.0.1:8080
```

### mitmproxy CA Certificate Installation
```bash
# Install mitmproxy CA certificate
# macOS
sudo security add-trusted-cert -d -r trustRoot \
  -k /Library/Keychains/System.keychain \
  ~/.mitmproxy/mitmproxy-ca-cert.pem

# Linux
sudo cp ~/.mitmproxy/mitmproxy-ca-cert.pem \
  /usr/local/share/ca-certificates/mitmproxy.crt
sudo update-ca-certificates
```

---

## 5. Backup and Key Rotation

### Key Rotation Principles

- **Personal API Keys**: Rotate every 60 days minimum
- **SSH Keys**: Rotate annually or on suspicion of compromise
- **System API Keys**: Use automated rotation (AWS Secrets Manager, HashiCorp Vault)
- **Passwords**: Follow organizational policy (typically quarterly)

### Backup Strategy

- **YubiKey**: Maintain 2+1 strategy (Primary + Secondary + Emergency offline)
- **SSH Keys**: Encrypted backup in password manager
- **Recovery Keys**: Store offline in secure physical location (never digitally)
- **2FA Backup Codes**: Print and store securely, never in cloud

### Example: Simple Rotation Reminder

```bash
# Add to ~/.zshrc or ~/.bashrc
# Check last rotation date on login
if [ -f ~/.security/last-rotation ]; then
    LAST_ROTATION=$(cat ~/.security/last-rotation)
    DAYS_SINCE=$((( $(date +%s) - $(date -d "$LAST_ROTATION" +%s) ) / 86400))
    if [ $DAYS_SINCE -gt 60 ]; then
        echo "⚠️  Keys last rotated $DAYS_SINCE days ago. Time to rotate!"
    fi
fi

# Mark rotation date
echo $(date +%Y-%m-%d) > ~/.security/last-rotation
```

---

## Security Checklist

### Essential (Required)
- [ ] Full‑disk encryption enabled
- [ ] All secrets removed from files
- [ ] OS keychain configured for credentials
- [ ] Biometric/hardware auth enabled
- [ ] Screen lock configured (5 min timeout)
- [ ] Production data purged after use

### Advanced (Recommended)
- [ ] Backup keys secured (2+1 strategy)
- [ ] Rotation reminders configured
- [ ] Network proxy monitoring (optional)
- [ ] Recovery keys stored offline

---

## Emergency Procedures

### Lost Hardware Token
- Switch to backup immediately
- Revoke lost device's certificates
- Rotate affected credentials
- Order replacement

### Suspected Compromise
**Immediate Actions (First 30 minutes):**
- Rotate ALL credentials
- Revoke active sessions
- Notify security team

**Follow-up Actions:**
- Review audit logs
- Re-encrypt disk
- Generate new SSH keys
- Document incident

---

## 6. Browser Security

### Common Attack Vectors
- **Malicious Extensions**: Can access all browsing data and inject code
- **Session Hijacking**: Cookies/tokens stolen via XSS or compromised extensions
- **Clipboard Access**: Extensions/sites can read clipboard containing secrets
- **Download Attacks**: Malware disguised as legitimate downloads

### Hardening Measures

**Profile Isolation:**
- Separate profiles for work vs personal browsing
- Use different browsers for sensitive operations (banking, cloud consoles)
- Incognito/Private mode for one-off sensitive tasks

**Extension Security:**
- Audit all extensions - remove unnecessary ones
- Check extension permissions before installing
- Prefer open-source extensions with active maintenance
- Disable extensions on sensitive sites

**Cookie & Session Management:**
- Enable "Clear cookies on exit" for non-essential sites
- Use password manager's browser extension (avoid browser's built-in)
- Enable 2FA/MFA on all critical accounts
- Log out of cloud consoles when not in use

**Advanced Protection:**
- Use Firefox containers or Chrome profiles for client isolation
- Enable browser's enhanced tracking protection
- Consider using a security-focused browser (Brave, hardened Firefox)
- Disable JavaScript on high-risk sites

---

## 7. Development Containers Security

### Common Mistakes

**Secret Leakage:**
- ❌ Hardcoding secrets in Dockerfiles or docker-compose.yml
- ❌ Building secrets into images (they persist in layers)
- ❌ Mounting ~/.aws or ~/.ssh directly into containers
- ✅ Use BuildKit secrets, environment variables at runtime, or secret management tools

**Excessive Permissions:**
- ❌ Running containers as root
- ❌ Using --privileged flag unnecessarily
- ❌ Mounting Docker socket (/var/run/docker.sock)
- ✅ Use USER directive, drop capabilities, run rootless containers

**Network Exposure:**
- ❌ Binding to 0.0.0.0 for local development
- ❌ Exposing unnecessary ports
- ✅ Bind to 127.0.0.1, use minimal port exposure

### Security Checklist for Dev Containers

- [ ] No secrets in Dockerfiles, compose files, or environment files
- [ ] Base images from trusted registries only
- [ ] Regular base image updates (security patches)
- [ ] Non-root user configured
- [ ] Minimal required capabilities
- [ ] Volume mounts limited to project directories
- [ ] Network isolation between containers
- [ ] Container logs don't contain sensitive data

### Quick Hardening

```bash
# Instead of mounting entire home directory
# ❌ docker run -v ~:/home/user image
# ✅ docker run -v $(pwd):/workspace image

# Instead of hardcoded secrets
# ❌ ENV API_KEY=secret123
# ✅ docker run --env-file .env.local image

# Instead of root user
# ❌ (default behavior)
# ✅ USER 1000:1000 (in Dockerfile)
```

---

## 8. Supply Chain Hygiene

### Package Managers
- Node.js: Use `npm ci`/`pnpm install --frozen-lockfile` with committed lockfiles; consider `npm config set ignore-scripts true` by default, enabling only when required.
- Python: Pin with hashes (e.g., `pip-compile --generate-hashes`), install with `pip install --require-hashes -r requirements.txt` inside a venv.
- Go: Commit `go.sum`, run `go mod verify`; set `GOPRIVATE` for internal modules.
- Rust: Commit `Cargo.lock`; consider `cargo deny` for advisories.

### Artifacts and Images
- Pin container base images by digest (`FROM ubuntu@sha256:…`); prefer minimal bases and `docker build --pull` for fresh layers.
- Verify signed images/binaries when available (Sigstore `cosign verify`, vendor signatures).

### Git Hygiene
- Require signed commits/tags; enable locally: `git config commit.gpgsign true` (or SSH signing).
- Review dependency diffs on bumps; avoid auto‑merging lockfile changes without review.

### Execution Hygiene
- Avoid piping curl to shell; prefer package managers.
- Run untrusted tools in throwaway VMs/containers; never with `--privileged` or host mounts.

---

## Resources

### Official Documentation
- [NIST Authentication Guidelines](https://pages.nist.gov/800-63-3/sp800-63b.html)
- [YubiKey Best Practices](https://developers.yubico.com/PIV/Guides/Best_practices.html)
- Platform Security Guides: [macOS](https://support.apple.com/guide/security/welcome/web) | [Linux](https://www.kernel.org/doc/html/latest/admin-guide/LSM/index.html)

### Quick References
- Rotation Schedule: See [Security Controls Table](#security-controls-summary-table)
