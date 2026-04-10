# Ravel Features

This document provides detailed information about Ravel's features and how to use them.

## Table of Contents

1. [Volumes](#volumes)
2. [Health Checks](#health-checks)
3. [Private Networks](#private-networks)
4. [Secrets Management](#secrets-management)
5. [Machine Configuration](#machine-configuration)

## Volumes

Ravel supports mounting persistent volumes to your machines. Volumes are backed by disks and can be attached to machines at specified mount paths.

### Configuration

```json
{
  "workload": {
    "volumes": [
      {
        "name": "data-disk",
        "path": "/data"
      },
      {
        "name": "cache-disk",
        "path": "/cache"
      }
    ]
  }
}
```

### Validation Rules

- **Maximum**: 10 volumes per machine
- **Unique names**: Volume names must be unique within a machine
- **Unique paths**: Mount paths must be unique within a machine
- **Absolute paths**: All mount paths must be absolute (start with `/`)
- **Non-empty**: Names and paths cannot be empty

### Example

See [volumes-example.json](examples/volumes-example.json) for a complete example.

### Disk Management

Before attaching a volume, you need to create a disk:

```bash
# Create a disk
curl -X POST http://localhost:3000/api/v1/namespaces/default/disks \
  -H "Content-Type: application/json" \
  -d '{
    "id": "data-disk",
    "size_mb": 10240
  }'

# List disks
curl http://localhost:3000/api/v1/namespaces/default/disks

# Delete a disk
curl -X DELETE http://localhost:3000/api/v1/namespaces/default/disks/data-disk
```

---

## Health Checks

Health checks allow you to monitor the health of your machines by running periodic commands inside the machine.

### Configuration

```json
{
  "workload": {
    "health_check": {
      "exec": ["wget", "--spider", "-q", "http://localhost:8080/health"],
      "interval": 10,
      "timeout": 3,
      "retries": 3
    }
  }
}
```

### Parameters

- **exec** (required): Command to run for health check
  - Array of strings representing the command and its arguments
  - Exit code 0 indicates healthy, non-zero indicates unhealthy

- **interval** (optional): Interval between health checks in seconds
  - Default: 30 seconds
  - Minimum: 1 second

- **timeout** (optional): Timeout for health check command in seconds
  - Default: 5 seconds
  - Command is killed if it exceeds this duration

- **retries** (optional): Number of consecutive failures before marking unhealthy
  - Default: 3
  - Machine is marked unhealthy after this many consecutive failures

### Health Status

Machines can have one of four health statuses:

- **unknown**: No health check configured or machine not yet started
- **starting**: Machine is starting, health checks not yet running
- **healthy**: Health check passing
- **unhealthy**: Health check failing for `retries` consecutive times

### Viewing Health Status

```bash
# Get machine details including health status
curl http://localhost:3000/api/v1/namespaces/default/fleets/my-fleet/machines/machine-id

# Response includes:
{
  "id": "machine-id",
  "status": "running",
  "health": "healthy",
  ...
}
```

### Example

See [healthcheck-example.json](examples/healthcheck-example.json) for a complete example.

---

## Private Networks

Ravel supports Wireguard-based encrypted private networks for secure machine-to-machine communication.

### Configuration

```json
{
  "workload": {
    "private_networks": [
      {
        "name": "app-network",
        "ip": "10.0.1.2/24"
      },
      {
        "name": "db-network",
        "ip": "10.0.2.2/24"
      }
    ]
  }
}
```

### Parameters

- **name** (required): Name of the private network to join
  - Must be unique within the machine
  - Used to identify the network across the cluster

- **ip** (required): IP address for this machine in the private network
  - Must be in CIDR notation (e.g., "10.0.1.2/24")
  - Must be unique within the network
  - IP should be within the network's subnet range

### Validation Rules

- **Maximum**: 5 private networks per machine
- **Unique names**: Network names must be unique within a machine
- **CIDR notation**: IP addresses must include prefix length (e.g., /24)
- **Non-empty**: Names and IPs cannot be empty

### How It Works

1. When a machine joins a private network, Ravel:
   - Generates a Wireguard keypair for the machine
   - Creates a Wireguard interface (e.g., `wg0`)
   - Assigns the specified IP to the interface
   - Configures peers (other machines in the same network)

2. Machines in the same private network can communicate:
   - All traffic is encrypted using Wireguard
   - Direct peer-to-peer connections when possible
   - Works across different regions/nodes

3. Network isolation:
   - Machines in different private networks cannot communicate
   - Private network traffic is separate from public traffic

### Example Use Cases

**Application Tier Communication:**
```json
{
  "private_networks": [
    {"name": "app-tier", "ip": "10.0.1.10/24"}
  ]
}
```

**Multi-Tier Architecture:**
```json
{
  "private_networks": [
    {"name": "frontend-backend", "ip": "10.0.1.5/24"},
    {"name": "backend-database", "ip": "10.0.2.5/24"}
  ]
}
```

### Example

See [private-network.json](examples/private-network.json) for a complete example.

---

## Secrets Management

Ravel provides secure secrets management for injecting sensitive data into machines as environment variables.

### Creating Secrets

```bash
# Create a secret
curl -X POST http://localhost:3000/api/v1/namespaces/default/secrets \
  -H "Content-Type: application/json" \
  -d '{
    "name": "db-password",
    "value": "super-secret-password"
  }'
```

### Using Secrets in Machines

```json
{
  "workload": {
    "secrets": [
      {
        "name": "db-password",
        "env_var": "DATABASE_PASSWORD"
      },
      {
        "name": "api-key",
        "env_var": "API_KEY"
      }
    ]
  }
}
```

### Parameters

- **name** (required): Name of the secret in the namespace
  - Must exist before creating the machine

- **env_var** (required): Environment variable name to inject the secret into
  - The secret value will be available as this environment variable inside the machine

### Security Considerations

- Secrets are stored encrypted in the database
- Secrets are injected as environment variables during machine creation
- Secret values are not exposed in API responses
- Secrets are namespace-scoped

### Managing Secrets

```bash
# List secrets (values are not returned)
curl http://localhost:3000/api/v1/namespaces/default/secrets

# Update a secret
curl -X PUT http://localhost:3000/api/v1/namespaces/default/secrets/db-password \
  -H "Content-Type: application/json" \
  -d '{"value": "new-password"}'

# Delete a secret
curl -X DELETE http://localhost:3000/api/v1/namespaces/default/secrets/db-password
```

---

## Machine Configuration

### Complete Configuration Example

Here's a complete machine configuration showcasing all features:

```json
{
  "image": "docker.io/library/nginx:latest",
  "guest": {
    "cpu_kind": "std",
    "cpus": 2,
    "memory_mb": 2048
  },
  "workload": {
    "init": {
      "cmd": ["/docker-entrypoint.sh", "nginx", "-g", "daemon off;"],
      "user": "root"
    },
    "env": [
      "NGINX_HOST=example.com",
      "NGINX_PORT=80"
    ],
    "secrets": [
      {
        "name": "tls-cert",
        "env_var": "TLS_CERTIFICATE"
      }
    ],
    "volumes": [
      {
        "name": "nginx-data",
        "path": "/usr/share/nginx/html"
      }
    ],
    "health_check": {
      "exec": ["wget", "--spider", "-q", "http://localhost"],
      "interval": 10,
      "timeout": 3,
      "retries": 3
    },
    "private_networks": [
      {
        "name": "web-tier",
        "ip": "10.0.1.5/24"
      }
    ],
    "restart": {
      "policy": "on-failure",
      "max_retries": 3
    },
    "auto_destroy": false
  },
  "stop_config": {
    "timeout": 10,
    "signal": "SIGTERM"
  }
}
```

### Guest Configuration

- **cpu_kind**: CPU template to use (e.g., "std")
- **cpus**: Number of virtual CPUs (must match template)
- **memory_mb**: Memory in megabytes (must match template)

### Workload Configuration

- **init**: Command to run inside the machine
  - **cmd**: Command and arguments
  - **entrypoint**: Override image entrypoint
  - **user**: User to run as

- **env**: Environment variables as array of "KEY=VALUE" strings

- **restart**: Restart policy configuration
  - **policy**: "always", "on-failure", or "never"
  - **max_retries**: Maximum restart attempts (for "on-failure")

- **auto_destroy**: Automatically destroy machine after exit

### Stop Configuration

- **timeout**: Seconds to wait before force-killing (default: 5, max: 30)
- **signal**: Signal to send for graceful shutdown (default: "SIGTERM")

---

## Best Practices

### Volumes
- Use volumes for persistent data that must survive machine restarts
- Separate data, logs, and cache onto different volumes
- Size volumes appropriately for expected data growth

### Health Checks
- Always configure health checks for production workloads
- Use application-specific health endpoints when available
- Set appropriate intervals based on application characteristics
- Consider health check overhead on application performance

### Private Networks
- Use private networks for all internal service communication
- Assign static IPs within the private network range
- Document IP allocation to avoid conflicts
- Separate networks by security zones (frontend, backend, database)

### Secrets
- Never hardcode sensitive data in machine configurations
- Rotate secrets regularly
- Use descriptive secret names
- Delete unused secrets promptly

### Resource Allocation
- Start with minimal resources and scale up based on observed usage
- Monitor resource usage to optimize allocation
- Use appropriate CPU kinds for workload characteristics
