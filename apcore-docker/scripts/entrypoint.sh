#!/bin/sh

# Wait for PostgreSQL to be available
until pg_isready -h db -p 5432 -U "$POSTGRES_USER"; do
  echo "Waiting for PostgreSQL..."
  sleep 2
done

# Run database initialization script
if [ -f /scripts/init-db.sh ]; then
  echo "Initializing database..."
  /scripts/init-db.sh
fi

# Start the application
exec "$@"