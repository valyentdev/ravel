# Testing & Debugging Examples

## Exec into a running machine

Execute commands inside a running machine using the API:

```bash
curl -X POST http://localhost:8080/v1/namespaces/{namespace}/fleets/{fleet}/machines/{machine_id}/exec \
  -H "Content-Type: application/json" \
  -d '{
    "cmd": ["ls", "-la", "/"],
    "timeout_ms": 5000
  }'
```

Response:
```json
{
  "stdout": "total 64\ndrwxr-xr-x   1 root root ...",
  "stderr": "",
  "exit_code": 0
}
```

## Health Checks

Configure automatic health checks to monitor machine health:

```json
{
  "image": "docker.io/library/nginx:alpine",
  "guest": {
    "vcpus": 1,
    "cpus_mhz": 1000,
    "memory_mb": 512
  },
  "workload": {
    "health_check": {
      "exec": ["wget", "--spider", "-q", "http://localhost"],
      "interval": 10,
      "timeout": 3,
      "retries": 3
    }
  }
}
```

**Health Check Parameters:**
- `exec`: Command to run for health check
- `interval`: Seconds between checks (default: 30)
- `timeout`: Command timeout in seconds (default: 5)
- `retries`: Consecutive failures before unhealthy (default: 3)

**Health Status Values:**
- `unknown`: No health check configured
- `starting`: Initial state after machine starts
- `healthy`: Health check passing
- `unhealthy`: Health check failing

The machine's health status is available in the machine API response under the `health` field.
