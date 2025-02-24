# Cloudflare DNS Export Tool

A simple Go utility to export DNS records from Cloudflare domains for backup, review, or migration to other DNS providers.

## Features

- Export DNS records from one or multiple Cloudflare zones
- Generate both JSON and BIND zone file formats
- Automatically handle Cloudflare-specific settings
- Support for API Token or Email+API Key authentication
- Configure via config file or environment variables

## Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/cloudflare-dns-export.git
cd cloudflare-dns-export

# Install dependencies
go mod init cloudflare-export
go get github.com/cloudflare/cloudflare-go
go get gopkg.in/yaml.v3

# Build the binary (optional)
go build -o cf-zonebackup cf-zonebackup.go
```

## Usage

### Configuration

1. **Via YAML file** - Create `config.yaml`:

```yaml
# Authentication (use either api_token OR email+api_key)
api_token: "your-api-token"  # Recommended
# OR
email: "your-email@example.com"
api_key: "your-global-api-key"

# Single zone
zone_name: "example.com"
# OR multiple zones
zone_names:
  - "example.com"
  - "example.org"
```

2. **Via environment variables**:

```bash
# Authentication
export CLOUDFLARE_API_TOKEN="your-api-token"
# OR
export CLOUDFLARE_EMAIL="your-email@example.com"
export CLOUDFLARE_API_KEY="your-global-api-key"

# Single zone
export CLOUDFLARE_ZONE_NAME="example.com"
# OR multiple zones (comma-separated)
export CLOUDFLARE_ZONE_NAMES="example.com,example.org"
```

### Running the tool

```bash
go run cf-zonebackup.go
# OR if built
./cf-zonebackup
```

Exports will be saved to a `cloudflare-export` directory with timestamped filenames.

## API Token Permissions

Create an API token with these permissions:
- Zone:Zone:Read
- Zone:DNS:Read

## Troubleshooting

If you encounter "Invalid request headers" errors:
1. Verify your API token has correct permissions
2. Check that your token is still valid 
3. Ensure your zone exists in your Cloudflare account
4. Try creating a new API token

## License

MIT
