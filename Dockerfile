FROM postgres:15

# Set PostgreSQL environment variables
ENV POSTGRES_USER=apcore
ENV POSTGRES_PASSWORD=secret
ENV POSTGRES_DB=apcoredb

# Install necessary dependencies
RUN apt-get update && apt-get install -y \
    openssl \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Create necessary directories
RUN mkdir -p /root/.apcore/certs

# Generate SSL certificates
RUN openssl req -x509 -newkey rsa:4096 \
    -keyout /root/.apcore/certs/key.pem \
    -out /root/.apcore/certs/cert.pem \
    -days 365 -nodes \
    -subj "/CN=localhost"

# Create a directory for the example application
WORKDIR /app

# Copy the example application (assuming it's in the build context)
COPY ./example /app/example

# Create an entrypoint script
RUN echo '#!/bin/bash \n\
# Start PostgreSQL in the background \n\
docker-entrypoint.sh postgres & \n\
PG_PID=$! \n\
# Wait for PostgreSQL to be ready \n\
until pg_isready -h localhost -p 5432 -U apcore; do \n\
  echo "Waiting for PostgreSQL..." \n\
  sleep 2 \n\
done \n\
echo "PostgreSQL is ready!" \n\
# Configure example \n\
/app/example configure \n\
# Initialize database \n\
/app/example init-db \n\
# Initialize admin account \n\
/app/example init-admin \n\
# Start the server \n\
/app/example serve & \n\
APP_PID=$! \n\
# Monitor processes \n\
wait $PG_PID $APP_PID' > /app/entrypoint.sh && \
chmod +x /app/entrypoint.sh

# Expose PostgreSQL port
EXPOSE 5432
# Expose the application port (assuming 8080)
EXPOSE 8080

# Set the entrypoint
ENTRYPOINT ["/app/entrypoint.sh"]