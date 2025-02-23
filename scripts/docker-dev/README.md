# Docker Development Environment

Docker-based development environment with PostgreSQL and essential development tools.

## Setup

```bash
# Start environment
docker-compose up -d

# Set root password
docker exec -it dev-environment passwd root

# Access dev environment (choose one)
docker exec -it dev-environment /bin/bash  # Direct shell
ssh -p 2222 root@localhost                 # SSH access

# Access PostgreSQL (choose one)
psql -h localhost -p 5432 -U devuser -d devdb
docker exec -it dev-postgres psql -U devuser -d devdb
```

## Resource Management

```bash
docker-compose pause    # Pause environment
docker-compose unpause  # Resume environment
docker-compose stop    # Stop containers
docker-compose start   # Start containers
docker-compose down    # Remove containers
```

## Volumes

- `go-cache`: Go modules
- `npm-cache`: NPM packages
- `pip-cache`: Python packages
- `venv`: Python environment
- `postgres-data`: PostgreSQL data

## Maintenance

```bash
# Clean rebuild
docker-compose build --no-cache
docker-compose up -d

# Remove all data
docker-compose down -v
```