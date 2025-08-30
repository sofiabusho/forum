#!/bin/bash
# build.sh - build and run the forum project with Docker

# Stop and remove old containers/networks/volumes
docker compose down -v

# Build and start the new container in detached mode
docker compose up -d --build

# List all Docker images
docker images   

# Show running containers
docker compose ps

# Check for unused volumes
docker volume ls