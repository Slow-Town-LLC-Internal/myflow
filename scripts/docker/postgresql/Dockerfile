# scripts/docker/postgresql/Dockerfile
FROM postgres:15-alpine

ENV POSTGRES_USER=admin
ENV POSTGRES_PASSWORD=password
ENV POSTGRES_DB=testdb

COPY init.sql /docker-entrypoint-initdb.d/

EXPOSE 5432


