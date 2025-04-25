#!/bin/bash

set -e

# Wait for PostgreSQL to be available
until pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER"; do
  echo "Waiting for PostgreSQL to be ready..."
  sleep 2
done

# Initialize the database
echo "Initializing the database..."

# Create the database if it doesn't exist
psql -h "$DB_HOST" -U "$DB_USER" -c "CREATE DATABASE IF NOT EXISTS $DB_NAME;"

# Run any additional SQL scripts or commands here
# For example, to create tables or seed data
# psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" -f /path/to/your/init_script.sql

echo "Database initialization complete."