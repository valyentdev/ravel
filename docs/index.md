# Ravel Documentation

Welcome to the Ravel documentation! Learn how to deploy and manage containers as microVMs with Ravel.

## Getting Started

- **[Quickstart Guide](quickstart.md)** - Get up and running in minutes
- **[Installation](runtime-setup.md)** - Detailed installation instructions
- **[Configuration](config.md)** - Configure Ravel for your environment

## Core Concepts

- **[Architecture](architecture.md)** - Understanding Ravel's design
- **[Features](features.md)** - Complete feature reference
  - Volumes - Persistent storage
  - Health Checks - Monitor machine health
  - Private Networks - Encrypted Wireguard networking
  - Secrets Management - Secure configuration
  - Machine Configuration - Resource management

## API Reference

- **[API Reference](api-reference.md)** - Complete HTTP API documentation
  - Namespaces & Fleets
  - Machines & Lifecycle
  - Disks & Storage
  - Secrets & Security

## Operations

- **[Databases](databases.md)** - PostgreSQL setup
- **[Development](development.md)** - Local dev environment with Docker Compose
- **[Production Deployment](production.md)** - Best practices for production

## Examples

Browse the [examples directory](examples/) for configuration samples:
- [hello-world.json](examples/hello-world.json) - Basic machine
- [volumes-example.json](examples/volumes-example.json) - Persistent storage
- [healthcheck-example.json](examples/healthcheck-example.json) - Health monitoring
- [private-network.json](examples/private-network.json) - Encrypted networking

## Reference

- **[Containerd](containerd.md)** - Image management details
- **[Runtime Setup](runtime-setup.md)** - Runtime configuration

## Community

- **GitHub**: [github.com/alexisbouchez/ravel](https://github.com/alexisbouchez/ravel)
- **Discord**: [Join our Discord](https://discord.gg/ekrFAtS6Bj)
- **Issues**: [Report bugs](https://github.com/alexisbouchez/ravel/issues)