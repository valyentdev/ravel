#!/bin/bash
# Simple hello-world microVM example

API=${RAVEL_API:-http://localhost:3000}
NAMESPACE=${NAMESPACE:-default}
FLEET=${FLEET:-my-fleet}

curl -X POST "${API}/fleets/${FLEET}/machines?namespace=${NAMESPACE}" \
  -H 'Content-Type: application/json' \
  -d '{
    "region": "eu-west",
    "config": {
      "image": "alpine:latest",
      "guest": {
        "cpu_kind": "std",
        "cpus": 1,
        "memory_mb": 256
      },
      "workload": {
        "init": {
          "cmd": ["/bin/sh", "-c", "echo Hello from Ravel! && sleep 10"]
        }
      }
    }
  }'
