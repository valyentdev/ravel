# Ravel Development Environment

## Quick Start with Docker Compose

The docker-compose setup provides the core development dependencies.

### Start the development stack:

```bash
docker-compose up -d
```

This starts:
- **PostgreSQL** (port 5432) - Ravel state database
- **NATS** (ports 4222, 8222) - Cluster messaging

### Access the services:

- **NATS Monitoring**: http://localhost:8222

### Configure Ravel Server

Example `ravel.toml`:

```toml
[server]
postgres_url = "postgresql://ravel:ravel_dev_password@localhost:5432/ravel"

[server.api]
address = ":3000"

[nats]
url = "nats://localhost:4222"
```

### Run Ravel Server

```bash
go run cmd/ravel/ravel.go server -c ravel.toml
```

### Stop the stack:

```bash
docker-compose down
```

To remove volumes (data will be lost):

```bash
docker-compose down -v
```
