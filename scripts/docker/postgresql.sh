
# scripts/docker/postgresql.sh
#!/bin/bash
set -e

CONTAINER_NAME="postgres-test"
IMAGE_NAME="myflow-postgres"

# Build image
podman build -t ${IMAGE_NAME} scripts/docker/postgresql/

# Run container
podman run -d \
  --name ${CONTAINER_NAME} \
  -p 5432:5432 \
  -v postgres_data:/var/lib/postgresql/data \
  ${IMAGE_NAME}

echo "PostgreSQL is running on localhost:5432"
echo "User: admin"
echo "Password: password"
echo "Database: testdb"
