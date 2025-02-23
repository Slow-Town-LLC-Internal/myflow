# MyFlow - Development Workflow Tools

A collection of development workflow tools and documentation, focusing on Go utilities, development environment setup, and technical documentation.

## Components
- **doc-generator**: A static site generator for markdown documentation
- **cluster-switcher**: Kubernetes cluster management utility
- **terraform-gcp-keys**: GCP key management tool
- Development environment setup using Docker Compose

## Quick Start

### Documentation Site
```bash
cd cmd/doc-generator
go build
./doc-generator -root /path/to/repo
```

### Development Environment
```bash
cd scripts/docker-dev
docker-compose up -d
```

## Project Structure
```
myflow/
├── cmd/                    # Go utilities
│   ├── doc-generator/      # Static site generator
│   ├── cluster-switcher/   # K8s cluster management
│   └── terraform-gcp-keys/ # GCP key management
├── docs/                   # Documentation content
│   ├── about.md
│   └── worklogs/          # Work logs by date
├── scripts/               
│   └── docker-dev/        # Development environment setup
```

## Prerequisites
- Go >= 1.22
- Docker & Docker Compose
- Git

## Contributing
1. Clone the repository
2. Create a feature branch
3. Submit a pull request

## License
MIT License
