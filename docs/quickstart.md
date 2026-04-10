# Ravel Quickstart Guide

Get up and running with Ravel in minutes.

## Prerequisites

Before installing Ravel, ensure you have:

- **Linux System** (tested on Ubuntu 20.04+, Debian 11+)
- **KVM** enabled (`lsmod | grep kvm`)
- **TUN/TAP** kernel module enabled
- **Root access** for installation
- **Go 1.22+** (for building from source)
- **Cloud Hypervisor** binary

## Quick Install

### Option 1: Install from Release (Recommended)

```bash
# Download and install the latest release
curl -fsSL https://raw.githubusercontent.com/alexisbouchez/ravel/master/install.sh | sudo bash

# Verify installation
ravel version
```

### Option 2: Build from Source

```bash
# Clone the repository
git clone https://github.com/alexisbouchez/ravel.git
cd ravel

# Build all binaries
make build-all

# Install binaries
sudo make install
```

## Dependencies Setup

### 1. Install Cloud Hypervisor

```bash
# Download Cloud Hypervisor
wget https://github.com/cloud-hypervisor/cloud-hypervisor/releases/download/v38.0/cloud-hypervisor-static
chmod +x cloud-hypervisor-static
sudo mv cloud-hypervisor-static /opt/ravel/cloud-hypervisor

# Download Linux kernel
wget https://github.com/cloud-hypervisor/cloud-hypervisor/releases/download/v38.0/vmlinux
sudo mv vmlinux /opt/ravel/vmlinux.bin
```

### 2. Install Containerd

```bash
# Install containerd for image management
sudo apt-get update
sudo apt-get install -y containerd

# Start containerd
sudo systemctl enable containerd
sudo systemctl start containerd
```

### 3. Setup NATS (Optional - for clustering)

```bash
# Install NATS Server
wget https://github.com/nats-io/nats-server/releases/download/v2.10.7/nats-server-v2.10.7-linux-amd64.tar.gz
tar -xzf nats-server-v2.10.7-linux-amd64.tar.gz
sudo mv nats-server-v2.10.7-linux-amd64/nats-server /usr/local/bin/

# Start NATS
nats-server &
```

## Configuration

### Single Node Setup

Create a configuration file at `/etc/ravel/config.toml`:

```toml
[daemon]
database_path = "/var/lib/ravel/daemon.db"

[daemon.runtime]
cloud_hypervisor_binary = "/opt/ravel/cloud-hypervisor"
jailer_binary = "/opt/ravel/jailer"
init_binary = "/opt/ravel/initd"
linux_kernel = "/opt/ravel/vmlinux.bin"

[daemon.agent]
resources = { cpus_mhz = 8000, memory_mb = 8192 }
node_id = "node-1"
region = "local"
address = "127.0.0.1"
port = 8080

[server]
postgres_url = "postgres://user:pass@localhost:5432/ravel"

[server.api]
address = ":3000"

[server.machine_templates.std]
vcpu_frequency = 2500
combinations = [
    { vcpus = 1, memory_configs = [512, 1024, 2048] },
    { vcpus = 2, memory_configs = [1024, 2048, 4096] },
]
```

### Create Required Directories

```bash
sudo mkdir -p /var/lib/ravel
sudo mkdir -p /opt/ravel
sudo mkdir -p /etc/ravel
```

## First Steps

### 1. Start Ravel Daemon

```bash
# Start the daemon (manages runtime and agent)
sudo ravel daemon -c /etc/ravel/config.toml
```

### 2. Start Ravel Server (in another terminal)

```bash
# Setup PostgreSQL database
createdb ravel
psql ravel < schema.sql

# Start the server (API and orchestrator)
ravel server -c /etc/ravel/config.toml
```

### 3. Create Your First Machine

```bash
# Create a namespace
curl -X POST http://localhost:3000/api/v1/namespaces \
  -H "Content-Type: application/json" \
  -d '{"name": "default"}'

# Create a fleet
curl -X POST http://localhost:3000/api/v1/namespaces/default/fleets \
  -H "Content-Type: application/json" \
  -d '{"name": "my-fleet"}'

# Create a machine
curl -X POST http://localhost:3000/api/v1/namespaces/default/fleets/my-fleet/machines \
  -H "Content-Type: application/json" \
  -d '{
    "region": "local",
    "config": {
      "image": "docker.io/library/alpine:latest",
      "guest": {
        "cpu_kind": "std",
        "cpus": 1,
        "memory_mb": 512
      },
      "workload": {
        "init": {
          "cmd": ["/bin/sh", "-c", "echo Hello from Ravel && sleep 3600"],
          "user": "root"
        }
      }
    }
  }'
```

### 4. Check Machine Status

```bash
# List machines in fleet
curl http://localhost:3000/api/v1/namespaces/default/fleets/my-fleet/machines

# Get specific machine
curl http://localhost:3000/api/v1/namespaces/default/fleets/my-fleet/machines/{machine-id}
```

### 5. Execute Command in Machine

```bash
curl -X POST http://localhost:3000/api/v1/namespaces/default/fleets/my-fleet/machines/{machine-id}/exec \
  -H "Content-Type: application/json" \
  -d '{
    "cmd": ["echo", "Hello from inside the machine"],
    "timeout_ms": 5000
  }'
```

## Next Steps

### Explore Features

- **[Volumes](features.md#volumes)**: Add persistent storage
- **[Health Checks](features.md#health-checks)**: Monitor machine health
- **[Private Networks](features.md#private-networks)**: Secure inter-machine communication
- **[Secrets](features.md#secrets-management)**: Manage sensitive data

### Examples

Check out the [examples directory](examples/) for common configurations:

- [hello-world.json](examples/hello-world.json) - Basic machine
- [volumes-example.json](examples/volumes-example.json) - With persistent storage
- [healthcheck-example.json](examples/healthcheck-example.json) - With health monitoring
- [private-network.json](examples/private-network.json) - With encrypted networking

### Production Setup

For production deployments:

1. **Enable TLS** for cluster communication
2. **Setup PostgreSQL** with replication
3. **Configure NATS** cluster
4. **Enable monitoring** and logging

See [Production Deployment](production.md) for detailed instructions.

## Troubleshooting

### Machine Won't Start

```bash
# Check daemon logs
sudo journalctl -u ravel-daemon -f

# Check machine events
curl http://localhost:3000/api/v1/namespaces/default/fleets/my-fleet/machines/{machine-id}
```

### Can't Connect to Machine

```bash
# Verify network configuration
ip addr show

# Check iptables rules
sudo iptables -L -n -v

# Test connectivity
ping {machine-ip}
```

### Image Pull Fails

```bash
# Verify containerd is running
sudo systemctl status containerd

# Check containerd logs
sudo journalctl -u containerd -f

# Manually pull image
sudo ctr image pull docker.io/library/alpine:latest
```

## Common Commands

```bash
# Build Ravel
make build-all

# Run tests
make test

# Start daemon in debug mode
sudo ravel daemon -c config.toml --debug

# View all machines across namespaces
curl http://localhost:3000/api/v1/machines

# Stop a machine
curl -X POST http://localhost:3000/api/v1/namespaces/default/fleets/my-fleet/machines/{machine-id}/stop

# Destroy a machine
curl -X DELETE http://localhost:3000/api/v1/namespaces/default/fleets/my-fleet/machines/{machine-id}
```

## Getting Help

- **Documentation**: [docs/](.)
- **GitHub Issues**: [github.com/alexisbouchez/ravel/issues](https://github.com/alexisbouchez/ravel/issues)
- **Discord**: [Join our Discord](https://discord.gg/ekrFAtS6Bj)

## What's Next?

- Read the [Architecture](architecture.md) guide to understand how Ravel works
- Learn about [Configuration](config.md) options
- Explore [Features](features.md) in depth
- Deploy to [Production](production.md)
