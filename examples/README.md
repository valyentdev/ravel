# Ravel Examples

Example scripts and configurations for running applications on Ravel.

## Prerequisites

- A running Ravel cluster with API accessible at `localhost:3000`
- A namespace called `default`
- A fleet called `my-fleet`

Create them if needed:
```bash
curl -X POST 'http://localhost:3000/namespaces' -H 'Content-Type: application/json' -d '{"name": "default"}'
curl -X POST 'http://localhost:3000/fleets?namespace=default' -H 'Content-Type: application/json' -d '{"name": "my-fleet"}'
```

## Examples

| Example | Description |
|---------|-------------|
| [hello-world](./hello-world/) | Simple Alpine container that prints a message |
| [nginx](./nginx/) | Nginx web server with gateway |
| [frankenphp](./frankenphp/) | Modern PHP application server |

## Creating a Gateway

To expose your application publicly, create a gateway:

```bash
curl -X POST 'http://localhost:3000/fleets/my-fleet/gateways?namespace=default' \
  -H 'Content-Type: application/json' \
  -d '{"name": "my-app", "target_port": 80}'
```

Your app will be available at `https://my-app.yourdomain.com`
