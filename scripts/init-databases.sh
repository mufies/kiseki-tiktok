#!/bin/bash
set -e

# Script to initialize multiple PostgreSQL databases for TikTok Clone microservices

echo "========================================="
echo "Initializing TikTok Clone Databases"
echo "========================================="

# Array of databases to create
databases=(
    "userdb"
    "videodb"
    "interactiondb"
    "eventdb"
    "feeddb"
    "notificationdb"
)

# Create each database if it doesn't exist
for db in "${databases[@]}"; do
    echo "Creating database: $db"
    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
        SELECT 'CREATE DATABASE $db'
        WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '$db')\gexec
EOSQL
    echo "✓ Database $db ready"
done

echo "========================================="
echo "All databases initialized successfully!"
echo "========================================="
echo ""
echo "Databases created:"
for db in "${databases[@]}"; do
    echo "  - $db"
done
echo ""
