# EKS Cluster Switcher

A simple command-line tool to switch between EKS clusters using AWS SSO profiles.

## Installation

```bash
# Clone the repository
git clone https://github.com/Slow-Town-LLC/cluster-switcher.git

# Install directly to your GOPATH
go install github.com/Slow-Town-LLC/cluster-switcher@latest
```

## Configuration

Create a configuration file at `~/.config/eks-config.yaml`:

```yaml
environments:
  prd:
    aws_profile: "production-profile"
    eks_cluster: "production-cluster"
    eks_region: "us-west-2"
  stg:
    aws_profile: "staging-profile"
    eks_cluster: "staging-cluster"
    eks_region: "us-west-2"
```

You can also specify a custom config path using the `EKS_CONFIG` environment variable:

```bash
export EKS_CONFIG=/path/to/your/config.yaml
```

## Usage

Switch to production cluster:
```bash
cluster-switcher prd
```

Switch to staging cluster:
```bash
cluster-switcher stg
```

## Features

- Automatic AWS SSO session verification and renewal
- Kubernetes context switching
- Environment variable management
- Configuration validation

## Requirements

- Go 1.19 or later
- AWS CLI v2
- kubectl
- Valid AWS SSO configuration

## Development

To build from source:

```bash
git clone https://github.com/yourusername/cluster-switcher.git
cd cluster-switcher
go build -o cluster-switcher
```

## License

MIT License
