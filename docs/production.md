# Production Deployment Guide

This guide covers best practices for deploying Ravel in production environments.

## Prerequisites

### System Requirements

**Minimum per Node:**
- 4 CPU cores
- 16 GB RAM
- 100 GB SSD storage
- Linux kernel 5.10+
- KVM support
- Network connectivity between nodes

**Recommended for Production:**
- 8+ CPU cores
- 32+ GB RAM
- 500+ GB SSD storage
- 10 Gbps network
- Dedicated network interfaces

### Software Requirements

- Go 1.22+ (for building)
- Cloud Hypervisor v38+
- Containerd 1.7+
- PostgreSQL 14+
- NATS Server 2.10+

## Architecture Overview

```
┌─────────────────────────────────────────────────────────┐
│                     Load Balancer                        │
│                   (HAProxy/nginx)                        │
└────────────────────┬────────────────────────────────────┘
                     │
        ┌────────────┴──────────────┐
        │                           │
┌───────▼────────┐        ┌────────▼────────┐
│  Ravel Server  │        │  Ravel Server   │
│   (Primary)    │◄──────►│   (Standby)     │
└───────┬────────┘        └────────┬────────┘
        │                          │
        └──────────┬───────────────┘
                   │
        ┌──────────▼──────────────┐
        │   PostgreSQL Cluster    │
        │  (Primary + Replicas)   │
        └──────────┬──────────────┘
                   │
        ┌──────────▼──────────────┐
        │     NATS Cluster        │
        │   (3+ nodes)            │
        └──────────┬──────────────┘
                   │
     ┌─────────────┼─────────────┐
     │             │             │
┌────▼───┐    ┌───▼────┐   ┌───▼────┐
│ Agent  │    │ Agent  │   │ Agent  │
│ Node 1 │    │ Node 2 │   │ Node 3 │
└────────┘    └────────┘   └────────┘
```

## Installation

### 1. Setup PostgreSQL Cluster

```bash
# On primary PostgreSQL server
sudo apt-get install postgresql-14

# Create database and user
sudo -u postgres psql
CREATE DATABASE ravel;
CREATE USER ravel WITH ENCRYPTED PASSWORD 'secure-password';
GRANT ALL PRIVILEGES ON DATABASE ravel TO ravel;

# Configure replication (postgresql.conf)
wal_level = replica
max_wal_senders = 3
max_replication_slots = 3
```

### 2. Setup NATS Cluster

```bash
# Install NATS Server
wget https://github.com/nats-io/nats-server/releases/download/v2.10.7/nats-server-v2.10.7-linux-amd64.tar.gz
tar -xzf nats-server-v2.10.7-linux-amd64.tar.gz
sudo mv nats-server-v2.10.7-linux-amd64/nats-server /usr/local/bin/

# Create NATS configuration (/etc/nats/nats.conf)
cluster {
  name: ravel-cluster
  listen: 0.0.0.0:6222
  routes: [
    nats://nats1:6222
    nats://nats2:6222
    nats://nats3:6222
  ]
}

# Start NATS with systemd
sudo systemctl enable nats
sudo systemctl start nats
```

### 3. Generate TLS Certificates

```bash
# Generate CA certificate
ravel tls ca \
  --output-dir ./certs \
  --organization "YourOrg" \
  --common-name "Ravel CA"

# Generate server certificates
ravel tls server \
  --ca-cert ./certs/ca-cert.pem \
  --ca-key ./certs/ca-key.pem \
  --output-dir ./certs \
  --server-id "server-1" \
  --ip "192.168.1.10"

# Generate agent certificates for each node
ravel tls agent \
  --ca-cert ./certs/ca-cert.pem \
  --ca-key ./certs/ca-key.pem \
  --output-dir ./certs \
  --agent-id "agent-1" \
  --ip "192.168.1.20"
```

### 4. Configure Ravel Server

Create `/etc/ravel/server.toml`:

```toml
[nats]
url = "nats://nats1:4222,nats://nats2:4222,nats://nats3:4222"
cred_file = "/etc/ravel/creds/nats.creds"

[server]
postgres_url = "postgres://ravel:password@postgres-primary:5432/ravel?sslmode=require"

[server.api]
address = ":3000"

[server.api.tls]
cert_file = "/etc/ravel/certs/server-cert.pem"
key_file = "/etc/ravel/certs/server-key.pem"
ca_file = "/etc/ravel/certs/ca-cert.pem"

[server.tls]
cert_file = "/etc/ravel/certs/server-cert.pem"
key_file = "/etc/ravel/certs/server-key.pem"
ca_file = "/etc/ravel/certs/ca-cert.pem"

[server.machine_templates.std]
vcpu_frequency = 2500
combinations = [
    { vcpus = 1, memory_configs = [512, 1024, 2048, 4096] },
    { vcpus = 2, memory_configs = [1024, 2048, 4096, 8192] },
    { vcpus = 4, memory_configs = [2048, 4096, 8192, 16384] },
    { vcpus = 8, memory_configs = [4096, 8192, 16384, 32768] },
]
```

### 5. Configure Ravel Agents

Create `/etc/ravel/agent.toml` on each agent node:

```toml
[nats]
url = "nats://nats1:4222,nats://nats2:4222,nats://nats3:4222"
cred_file = "/etc/ravel/creds/nats.creds"

[daemon]
database_path = "/var/lib/ravel/daemon.db"

[daemon.runtime]
cloud_hypervisor_binary = "/opt/ravel/cloud-hypervisor"
jailer_binary = "/opt/ravel/jailer"
init_binary = "/opt/ravel/initd"
linux_kernel = "/opt/ravel/vmlinux.bin"

[daemon.agent]
resources = { cpus_mhz = 20000, memory_mb = 28672 }
node_id = "agent-1"  # Unique per agent
region = "us-east-1"
address = "192.168.1.20"  # Public IP
port = 8080

[daemon.agent.tls]
cert_file = "/etc/ravel/certs/agent-cert.pem"
key_file = "/etc/ravel/certs/agent-key.pem"
ca_file = "/etc/ravel/certs/ca-cert.pem"
```

### 6. Setup Systemd Services

**Server service** (`/etc/systemd/system/ravel-server.service`):

```ini
[Unit]
Description=Ravel Server
After=network.target postgresql.service nats.service

[Service]
Type=simple
User=ravel
Group=ravel
ExecStart=/usr/bin/ravel server -c /etc/ravel/server.toml
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

**Agent service** (`/etc/systemd/system/ravel-agent.service`):

```ini
[Unit]
Description=Ravel Agent
After=network.target containerd.service

[Service]
Type=simple
User=root
ExecStart=/usr/bin/ravel daemon -c /etc/ravel/agent.toml
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

Enable and start services:

```bash
sudo systemctl enable ravel-server
sudo systemctl start ravel-server

sudo systemctl enable ravel-agent
sudo systemctl start ravel-agent
```

## Security Best Practices

### Network Security

1. **Firewall Configuration**
   ```bash
   # Server ports
   sudo ufw allow 3000/tcp    # API
   sudo ufw allow 4222/tcp    # NATS client
   sudo ufw allow 6222/tcp    # NATS cluster
   sudo ufw allow 8080/tcp    # Agent API
   sudo ufw allow 8082/tcp    # HTTP proxy

   # Enable firewall
   sudo ufw enable
   ```

2. **TLS Everywhere**
   - Enable TLS for all API communication
   - Use mutual TLS (mTLS) between components
   - Rotate certificates regularly

3. **Network Isolation**
   - Use VPCs or VLANs to isolate components
   - Restrict PostgreSQL to internal network
   - Use private networks for inter-machine communication

### Access Control

1. **API Authentication** (implement custom auth layer)
   - Use API tokens or OAuth2
   - Implement rate limiting
   - Log all API requests

2. **Namespace Isolation**
   - Use separate namespaces for different teams/apps
   - Implement RBAC for namespace access

3. **Secrets Management**
   - Rotate secrets regularly
   - Use strong encryption
   - Audit secret access

### System Security

1. **Kernel Security**
   ```bash
   # Enable kernel security features
   sudo sysctl -w kernel.unprivileged_userns_clone=0
   sudo sysctl -w kernel.kptr_restrict=2
   ```

2. **Container Security**
   - Run machines with minimal privileges
   - Use read-only filesystems where possible
   - Enable SELinux/AppArmor

## Monitoring

### Logging

```bash
# View server logs
journalctl -u ravel-server -f

# View agent logs
journalctl -u ravel-agent -f

# Filter by machine ID
journalctl -u ravel-agent -f | grep "machine_id=xyz"
```

### Health Checks

```bash
# Check server health
curl https://server:3000/health

# Check agent health
curl http://agent:8080/health

# Check machine health
curl https://server:3000/api/v1/namespaces/default/fleets/web/machines | \
  jq '.[] | select(.health != "healthy")'
```

## Backup and Recovery

### Database Backups

```bash
# Automated daily backups
0 2 * * * pg_dump -h localhost -U ravel ravel | gzip > /backup/ravel-$(date +\%Y\%m\%d).sql.gz

# Keep last 30 days
find /backup -name "ravel-*.sql.gz" -mtime +30 -delete
```

### Configuration Backups

```bash
# Backup configurations
tar -czf ravel-config-$(date +%Y%m%d).tar.gz \
  /etc/ravel \
  /etc/systemd/system/ravel-*.service
```

### Disaster Recovery

1. **Database Recovery**
   ```bash
   # Restore from backup
   gunzip < ravel-20240115.sql.gz | psql -h localhost -U ravel ravel
   ```

2. **Agent State Recovery**
   - Agents maintain local state in `/var/lib/ravel/daemon.db`
   - State is automatically synced to cluster on restart
   - Lost machines are automatically rescheduled

## Scaling

### Horizontal Scaling

**Add Server Nodes:**
1. Deploy new server with same configuration
2. Point to same PostgreSQL cluster
3. Add to load balancer pool

**Add Agent Nodes:**
1. Install Ravel on new node
2. Configure with unique `node_id`
3. Generate new TLS certificates
4. Start agent service

### Vertical Scaling

**Increase Agent Capacity:**
```toml
[daemon.agent]
resources = { cpus_mhz = 40000, memory_mb = 61440 }  # Doubled
```

**Database Scaling:**
- Use PostgreSQL read replicas
- Implement connection pooling (PgBouncer)
- Optimize queries and indexes

## Troubleshooting

### Common Issues

**Machines Won't Start:**
```bash
# Check agent logs
journalctl -u ravel-agent -n 100

# Verify Cloud Hypervisor
/opt/ravel/cloud-hypervisor --version

# Check KVM
lsmod | grep kvm
```

**Cluster Communication Issues:**
```bash
# Test NATS connectivity
nats-server --signal check

# Verify TLS certificates
openssl verify -CAfile ca-cert.pem agent-cert.pem
```

**High Resource Usage:**
```bash
# List machines by CPU usage
curl https://server:3000/api/v1/machines | \
  jq '.[] | select(.status == "running")' | \
  jq -r '.config.guest | "\(.cpus) \(.memory_mb)"'
```

## Performance Tuning

### System Tuning

```bash
# Increase file descriptor limits
echo "* soft nofile 65536" >> /etc/security/limits.conf
echo "* hard nofile 65536" >> /etc/security/limits.conf

# Optimize network stack
sysctl -w net.core.rmem_max=16777216
sysctl -w net.core.wmem_max=16777216
sysctl -w net.ipv4.tcp_rmem="4096 87380 16777216"
sysctl -w net.ipv4.tcp_wmem="4096 65536 16777216"
```

### Database Tuning

```sql
-- Optimize PostgreSQL settings
ALTER SYSTEM SET shared_buffers = '4GB';
ALTER SYSTEM SET effective_cache_size = '12GB';
ALTER SYSTEM SET work_mem = '64MB';
ALTER SYSTEM SET maintenance_work_mem = '512MB';
```

## Maintenance

### Rolling Updates

1. **Update Servers:**
   ```bash
   # Update server one at a time
   systemctl stop ravel-server
   cp /path/to/new/ravel /usr/bin/ravel
   systemctl start ravel-server
   ```

2. **Update Agents:**
   ```bash
   # Drain node
   curl -X POST https://server:3000/api/v1/nodes/agent-1/drain

   # Wait for machines to migrate
   watch 'curl https://server:3000/api/v1/nodes/agent-1 | jq .machine_count'

   # Update agent
   systemctl stop ravel-agent
   cp /path/to/new/ravel /usr/bin/ravel
   systemctl start ravel-agent

   # Uncordon node
   curl -X POST https://server:3000/api/v1/nodes/agent-1/uncordon
   ```

### Certificate Rotation

```bash
# Generate new certificates
ravel tls agent --ca-cert ca-cert.pem --ca-key ca-key.pem \
  --output-dir /etc/ravel/certs-new --agent-id agent-1

# Update configuration to use new certs
# Restart services with zero-downtime reload
systemctl reload ravel-agent
```

## Cost Optimization

1. **Right-Size Machines:** Monitor actual resource usage and adjust
2. **Use Auto-Destroy:** Enable for ephemeral workloads
3. **Schedule Non-Critical Workloads:** Run during off-peak hours
4. **Pool Resources:** Share agent nodes across namespaces

## Support

For production support:
- GitHub Issues: https://github.com/alexisbouchez/ravel/issues
- Discord: https://discord.gg/ekrFAtS6Bj
