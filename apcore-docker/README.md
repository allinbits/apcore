# apcore-docker

## Overview

This project sets up a PostgreSQL database and runs the example application using Docker. It utilizes Docker Compose to manage the services and provides a streamlined way to develop and test the application.

## Project Structure

- **docker-compose.yml**: Defines the services for the application, including the PostgreSQL database and the application itself.
- **Dockerfile**: Contains instructions to build the application image, setting up the necessary environment and dependencies.
- **.env**: Holds environment variable definitions used in the Docker Compose configuration.
- **config/config.ini.template**: A template for the application's configuration file, outlining the structure and default values.
- **scripts/entrypoint.sh**: Script executed when the Docker container starts, responsible for environment setup and application launch.
- **scripts/init-db.sh**: Initializes the PostgreSQL database, creating tables and seeding data as needed.

## Getting Started

### Prerequisites

- Docker
- Docker Compose

### Setup

1. Clone the repository:
   ```
   git clone <repository-url>
   cd apcore-docker
   ```

2. Create a `.env` file based on the provided template:
   ```
   cp .env.example .env
   ```

3. Modify the `.env` file to set your database credentials and other configurations.

### Running the Application

1. Start the services using Docker Compose:
   ```
   docker-compose up
   ```

2. Access the application at `http://localhost:YOUR_PORT`.

### Stopping the Application

To stop the services, run:
```
docker-compose down
```

### Database Initialization

The database will be initialized automatically when the application starts. You can customize the initialization process by modifying the `scripts/init-db.sh` script.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any improvements or bug fixes.

## License

This project is licensed under the MIT License. See the LICENSE file for details.