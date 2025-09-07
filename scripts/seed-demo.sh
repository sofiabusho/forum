#!/bin/bash

echo "Loading demo data into forum database..."

# Check if running in Docker or locally
if [ -f "/app/data/forum.db" ]; then
    # Running in Docker
    DB_PATH="/app/data/forum.db"
else
    # Running locally
    DB_PATH="./forum.db"
fi
# Load the seed data
sqlite3 "$DB_PATH" < /app/scripts/seed-demo-data.sql
