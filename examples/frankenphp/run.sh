#!/bin/bash
# FrankenPHP application server example

API=${RAVEL_API:-http://localhost:3000}
NAMESPACE=${NAMESPACE:-default}
FLEET=${FLEET:-my-fleet}
GATEWAY_NAME=${GATEWAY_NAME:-php}

echo "Creating FrankenPHP machine..."
RESPONSE=$(curl -s -X POST "${API}/fleets/${FLEET}/machines?namespace=${NAMESPACE}" \
  -H 'Content-Type: application/json' \
  -d '{
    "region": "eu-west",
    "config": {
      "image": "dunglas/frankenphp:latest",
      "guest": {
        "cpu_kind": "std",
        "cpus": 1,
        "memory_mb": 512
      },
      "workload": {
        "env": ["SERVER_NAME=:80"],
        "restart": {
          "policy": "always"
        }
      }
    }
  }')

MACHINE_ID=$(echo $RESPONSE | jq -r '.id')
echo "Machine created: $MACHINE_ID"

echo "Creating gateway..."
curl -s -X POST "${API}/fleets/${FLEET}/gateways?namespace=${NAMESPACE}" \
  -H 'Content-Type: application/json' \
  -d "{\"name\": \"${GATEWAY_NAME}\", \"target_port\": 80}"

echo ""
echo "FrankenPHP will be available at: https://${GATEWAY_NAME}.yourdomain.com"
