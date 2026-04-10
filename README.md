# Ravel

> Run containers as microVMs. Fast, secure, simple.

<div align="center">
  <a href="https://github.com/alexisbouchez/ravel/actions/workflows/ci.yml">
    <img src="https://github.com/alexisbouchez/ravel/actions/workflows/ci.yml/badge.svg" alt="CI" />
  </a>
  <a href="https://discord.gg/ekrFAtS6Bj">
    <img src="https://img.shields.io/badge/chat-on%20discord-7289DA.svg" alt="Discord Chat" />
  </a>
  <a href="https://x.com/AlexisBouchezFR">
    <img src="https://img.shields.io/twitter/follow/AlexisBouchezFR.svg?label=Follow%20@AlexisBouchezFR" alt="Follow @AlexisBouchezFR" />
  </a>
</div>

## What is Ravel?

Ravel is an open-source orchestrator that turns OCI images inside lightweight microVMs powered by [Cloud Hypervisor](https://github.com/cloud-hypervisor/cloud-hypervisor). Get the sandboxing of VMs with the simplicity of containers.

```bash
# Deploy nginx in a microVM
curl -X POST 'http://localhost:3000/fleets/my-fleet/machines?namespace=default' \
  -H 'Content-Type: application/json' \
  -d '{"region": "eu-west", "config": {"image": "nginx:alpine", "guest": {"cpus": 1, "memory_mb": 512}}}'
```

## Table of Contents

- [Features](#features)
- [Quick Start](#quick-start)
- [Examples](#examples)
- [Architecture](#architecture)
- [Documentation](#documentation)
- [FAQ](#faq)
- [License](#license)

## Features

- **MicroVM Runtime** - Run any OCI/Docker image inside isolated Cloud Hypervisor VMs
- **HTTP Gateway** - Automatic HTTPS routing with Let's Encrypt support
- **TCP/UDP Proxy** - Load-balanced proxy for any protocol
- **Private Networks** - WireGuard-based networking between machines
- **Volumes & Secrets** - Persistent storage and secure secret management
- **Clustering** - Multi-node clusters with gossip-based state sync

## Quick Start

```bash
# 1. Create a namespace
curl -X POST 'http://localhost:3000/namespaces' \
  -H 'Content-Type: application/json' -d '{"name": "default"}'

# 2. Create a fleet
curl -X POST 'http://localhost:3000/fleets?namespace=default' \
  -H 'Content-Type: application/json' -d '{"name": "web"}'

# 3. Deploy a machine
curl -X POST 'http://localhost:3000/fleets/web/machines?namespace=default' \
  -H 'Content-Type: application/json' \
  -d '{"region": "eu-west", "config": {"image": "nginx:alpine", "guest": {"cpu_kind": "std", "cpus": 1, "memory_mb": 512}}}'

# 4. Create a gateway to expose it
curl -X POST 'http://localhost:3000/fleets/web/gateways?namespace=default' \
  -H 'Content-Type: application/json' -d '{"name": "www", "target_port": 80}'
```

Your app is now live at `https://www.yourdomain.com`

## Examples

| Example                                | Description                   |
| -------------------------------------- | ----------------------------- |
| [hello-world](./examples/hello-world/) | Simple Alpine container       |
| [nginx](./examples/nginx/)             | Nginx web server with gateway |
| [frankenphp](./examples/frankenphp/)   | Modern PHP application server |

See the [examples directory](./examples/) for more.

## Architecture

```
┌─────────────┐     ┌─────────────┐
│   Client    │────▶│   Ravel     │
└─────────────┘     └─────────────┘
                           │
                           ▼
                    ┌─────────────┐
                    │   MicroVM   │
                    │ (Container) │
                    └─────────────┘
```

**Components:**

- **Cloud Hypervisor** - Lightweight VMM for microVMs
- **Containerd** - OCI image management
- **NATS** - Pub/sub messaging

## Documentation

- [Installation Guide](./docs/index.md)
- [API Reference](./docs/api-reference.md)
- [Production Deployment](./docs/production.md)

## FAQ

### Is it production-ready?

Not yet, Ravel is in **beta** and is to be considered unstable.

But we are working full time on providing a stable release.

### Why is it named Ravel?

Ravel is named after the famous composer Maurice Ravel, known for his orchestral works.

### How do I contribute?

Please come and join us on our [Discord server](https://discord.gg/ekrFAtS6Bj), where you can ask questions, get help, and contribute to the project.

### How do I report a bug?

Please open an issue on our [GitHub repository](https://github.com/alexisbouchez/ravel/issues).

### How do I request a feature?

Please open an issue on our [GitHub repository](https://github.com/alexisbouchez/ravel/issues).

## License

Copyright 2026 Alexis Bouchez

This program is free software: you can redistribute it and/or modify it
under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or (at
your option) any later version.

This program is distributed in the hope that it will be useful, but
WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU Affero
General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see
[https://www.gnu.org/licenses/](https://www.gnu.org/licenses/).

Portions of this project (`pkg/vsock/`) are derived from the Firecracker
project, © Amazon.com, Inc., and remain under the Apache License 2.0 —
see the file headers for details.

## Star History

Thank you for your support! 🌟

[![Star History Chart](https://api.star-history.com/chart?repos=alexisbouchez/ravel&type=date&legend=top-left)](https://www.star-history.com/?repos=alexisbouchez%2Fravel&type=date&legend=top-left)
