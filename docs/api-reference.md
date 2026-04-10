# Ravel API Reference

Complete reference for the Ravel HTTP API.

## Base URL

```
http://localhost:3000/api/v1
```

## Authentication

Currently, Ravel API does not require authentication. TLS can be enabled for secure communication.

## Common Headers

```
Content-Type: application/json
Accept: application/json
```

---

## Namespaces

Namespaces provide logical isolation for resources.

### Create Namespace

```http
POST /namespaces
```

**Request Body:**
```json
{
  "name": "production"
}
```

**Response:** `201 Created`
```json
{
  "name": "production",
  "created_at": "2024-01-15T10:30:00Z"
}
```

### List Namespaces

```http
GET /namespaces
```

**Response:** `200 OK`
```json
[
  {
    "name": "default",
    "created_at": "2024-01-15T10:00:00Z"
  },
  {
    "name": "production",
    "created_at": "2024-01-15T10:30:00Z"
  }
]
```

### Get Namespace

```http
GET /namespaces/{namespace}
```

**Response:** `200 OK`
```json
{
  "name": "production",
  "created_at": "2024-01-15T10:30:00Z"
}
```

### Delete Namespace

```http
DELETE /namespaces/{namespace}
```

**Response:** `204 No Content`

---

## Fleets

Fleets are groups of machines with shared configuration.

### Create Fleet

```http
POST /namespaces/{namespace}/fleets
```

**Request Body:**
```json
{
  "name": "web-servers",
  "metadata": {
    "labels": {
      "tier": "frontend",
      "env": "production"
    },
    "annotations": {
      "description": "Production web servers"
    }
  }
}
```

**Response:** `201 Created`
```json
{
  "id": "fleet_abc123",
  "name": "web-servers",
  "namespace": "production",
  "metadata": { ... },
  "created_at": "2024-01-15T11:00:00Z"
}
```

### List Fleets

```http
GET /namespaces/{namespace}/fleets
```

**Response:** `200 OK`
```json
[
  {
    "id": "fleet_abc123",
    "name": "web-servers",
    "namespace": "production",
    "created_at": "2024-01-15T11:00:00Z"
  }
]
```

### Get Fleet

```http
GET /namespaces/{namespace}/fleets/{fleet}
```

**Response:** `200 OK`

### Update Fleet Metadata

```http
PUT /namespaces/{namespace}/fleets/{fleet}/metadata
```

**Request Body:**
```json
{
  "labels": {
    "tier": "frontend",
    "env": "staging"
  },
  "annotations": {
    "updated": "2024-01-15"
  }
}
```

**Response:** `200 OK`

### Delete Fleet

```http
DELETE /namespaces/{namespace}/fleets/{fleet}
```

**Response:** `204 No Content`

---

## Machines

Machines are individual VM instances running workloads.

### Create Machine

```http
POST /namespaces/{namespace}/fleets/{fleet}/machines
```

**Request Body:**
```json
{
  "region": "us-east-1",
  "config": {
    "image": "docker.io/library/nginx:latest",
    "guest": {
      "cpu_kind": "std",
      "cpus": 2,
      "memory_mb": 2048
    },
    "workload": {
      "init": {
        "cmd": ["nginx", "-g", "daemon off;"],
        "user": "root"
      },
      "env": ["NGINX_PORT=80"],
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
  },
  "skip_start": false,
  "enable_machine_gateway": false,
  "metadata": {
    "labels": {
      "app": "nginx",
      "version": "1.25"
    }
  }
}
```

**Response:** `201 Created`
```json
{
  "id": "machine_xyz789",
  "namespace": "production",
  "fleet": "web-servers",
  "instance_id": "inst_123",
  "machine_version": "v1",
  "region": "us-east-1",
  "config": { ... },
  "status": "starting",
  "health": "starting",
  "gateway_enabled": false,
  "metadata": { ... },
  "created_at": "2024-01-15T12:00:00Z",
  "updated_at": "2024-01-15T12:00:00Z",
  "events": []
}
```

### List Machines

```http
GET /namespaces/{namespace}/fleets/{fleet}/machines
```

**Query Parameters:**
- `status` - Filter by status (created, starting, running, stopped, etc.)
- `health` - Filter by health status (unknown, healthy, unhealthy)

**Response:** `200 OK`
```json
[
  {
    "id": "machine_xyz789",
    "status": "running",
    "health": "healthy",
    ...
  }
]
```

### Get Machine

```http
GET /namespaces/{namespace}/fleets/{fleet}/machines/{machine}
```

**Response:** `200 OK`
```json
{
  "id": "machine_xyz789",
  "namespace": "production",
  "fleet": "web-servers",
  "status": "running",
  "health": "healthy",
  "config": { ... },
  "events": [
    {
      "id": "event_001",
      "type": "machine.started",
      "timestamp": "2024-01-15T12:00:30Z",
      "payload": {
        "started": {
          "started_at": "2024-01-15T12:00:30Z"
        }
      }
    }
  ],
  ...
}
```

### Start Machine

```http
POST /namespaces/{namespace}/fleets/{fleet}/machines/{machine}/start
```

**Request Body (optional):**
```json
{
  "is_restart": false
}
```

**Response:** `200 OK`

### Stop Machine

```http
POST /namespaces/{namespace}/fleets/{fleet}/machines/{machine}/stop
```

**Request Body (optional):**
```json
{
  "timeout": 10,
  "signal": "SIGTERM"
}
```

**Response:** `200 OK`

### Execute Command

```http
POST /namespaces/{namespace}/fleets/{fleet}/machines/{machine}/exec
```

**Request Body:**
```json
{
  "cmd": ["ls", "-la", "/var/log"],
  "timeout_ms": 5000
}
```

**Response:** `200 OK`
```json
{
  "stdout": "total 24\ndrwxr-xr-x 3 root root 4096 Jan 15 12:00 .\n...",
  "stderr": "",
  "exit_code": 0
}
```

### Update Machine Metadata

```http
PUT /namespaces/{namespace}/fleets/{fleet}/machines/{machine}/metadata
```

**Request Body:**
```json
{
  "labels": {
    "app": "nginx",
    "version": "1.26"
  },
  "annotations": {
    "last_updated": "2024-01-15"
  }
}
```

**Response:** `200 OK`

### Delete Machine

```http
DELETE /namespaces/{namespace}/fleets/{fleet}/machines/{machine}
```

**Query Parameters:**
- `force` - Force delete even if running (default: false)

**Response:** `204 No Content`

---

## Disks

Persistent storage volumes that can be attached to machines.

### Create Disk

```http
POST /namespaces/{namespace}/disks
```

**Request Body:**
```json
{
  "id": "data-disk-001",
  "size_mb": 10240
}
```

**Response:** `201 Created`
```json
{
  "id": "data-disk-001",
  "namespace": "production",
  "size_mb": 10240,
  "attached_to": null,
  "created_at": "2024-01-15T13:00:00Z"
}
```

### List Disks

```http
GET /namespaces/{namespace}/disks
```

**Response:** `200 OK`
```json
[
  {
    "id": "data-disk-001",
    "size_mb": 10240,
    "attached_to": "machine_xyz789",
    "created_at": "2024-01-15T13:00:00Z"
  }
]
```

### Get Disk

```http
GET /namespaces/{namespace}/disks/{disk}
```

**Response:** `200 OK`

### Delete Disk

```http
DELETE /namespaces/{namespace}/disks/{disk}
```

**Response:** `204 No Content`

**Note:** Disk must not be attached to any machine.

---

## Secrets

Secure storage for sensitive configuration data.

### Create Secret

```http
POST /namespaces/{namespace}/secrets
```

**Request Body:**
```json
{
  "name": "database-password",
  "value": "super-secret-password-123"
}
```

**Response:** `201 Created`
```json
{
  "name": "database-password",
  "namespace": "production",
  "created_at": "2024-01-15T14:00:00Z"
}
```

**Note:** The `value` field is never returned in GET requests.

### List Secrets

```http
GET /namespaces/{namespace}/secrets
```

**Response:** `200 OK`
```json
[
  {
    "name": "database-password",
    "namespace": "production",
    "created_at": "2024-01-15T14:00:00Z"
  },
  {
    "name": "api-key",
    "namespace": "production",
    "created_at": "2024-01-15T14:10:00Z"
  }
]
```

### Get Secret

```http
GET /namespaces/{namespace}/secrets/{secret}
```

**Response:** `200 OK`
```json
{
  "name": "database-password",
  "namespace": "production",
  "created_at": "2024-01-15T14:00:00Z"
}
```

**Note:** Secret value is not included in response for security.

### Update Secret

```http
PUT /namespaces/{namespace}/secrets/{secret}
```

**Request Body:**
```json
{
  "value": "new-password-456"
}
```

**Response:** `200 OK`

### Delete Secret

```http
DELETE /namespaces/{namespace}/secrets/{secret}
```

**Response:** `204 No Content`

---

## Machine States

Machines transition through the following states:

| State | Description |
|-------|-------------|
| `created` | Machine record created, not yet scheduled |
| `preparing` | Machine being prepared on agent node |
| `starting` | Machine VM is starting |
| `running` | Machine is running |
| `stopping` | Machine is being stopped |
| `stopped` | Machine has stopped |
| `destroying` | Machine is being destroyed |
| `destroyed` | Machine has been destroyed |

## Health States

| State | Description |
|-------|-------------|
| `unknown` | No health check configured or insufficient data |
| `starting` | Machine starting, health checks not yet active |
| `healthy` | Health checks passing |
| `unhealthy` | Health checks failing |

## Error Responses

All error responses follow this format:

```json
{
  "error": {
    "code": "INVALID_ARGUMENT",
    "message": "Volume name cannot be empty",
    "details": {}
  }
}
```

### Common Error Codes

| HTTP Status | Code | Description |
|-------------|------|-------------|
| 400 | `INVALID_ARGUMENT` | Invalid request parameters |
| 404 | `NOT_FOUND` | Resource not found |
| 409 | `ALREADY_EXISTS` | Resource already exists |
| 409 | `CONFLICT` | Operation conflicts with current state |
| 500 | `INTERNAL_ERROR` | Internal server error |
| 503 | `UNAVAILABLE` | Service temporarily unavailable |

## Rate Limiting

Currently no rate limiting is enforced. This may change in future versions.

## Pagination

List endpoints do not currently support pagination. This may be added in future versions.

## Filtering and Sorting

Limited filtering is available on some list endpoints via query parameters. Sorting is not currently supported.

## Webhooks

Webhooks are not currently supported but are planned for a future release.

## Best Practices

1. **Always check machine events** when operations fail
2. **Use health checks** for production workloads
3. **Handle transient states** (preparing, starting, stopping)
4. **Implement retries** with exponential backoff
5. **Set appropriate timeouts** for exec commands
6. **Clean up resources** when no longer needed

## Examples

See the [examples directory](examples/) for complete configuration examples and the [features guide](features.md) for detailed feature documentation.
