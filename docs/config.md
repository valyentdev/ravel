# Ravel configuration

Ravel accept one unique TOML configuration file.
This file is usually located at `/etc/ravel/config.toml`.

## Common configuration

To run in a cluster configuration, you need to give an access to a [NATS server](https://docs.nats.io/running-a-nats-service/introduction) and a [Corrosion](./databases.md) node both to the Agent and the Server.

```toml
[nats]
url = "nats://127.0.0.1:4222"
cred_file = "./nats.creds"

[corrosion]
url = "http://127.0.0.1:8081"
pg_wire_addr = "postgres://127.0.0.1:5432"
```

## Daemon configuration

The Daemon is the process responsible to manage the Ravel Runtime and the  Ravel Agent.The daemon holds its state in a [bbolt](https://github.com/etcd-io/bbolt).

```toml
[daemon]
database_path = "/var/lib/ravel/daemon.db"
```

### Runtime configuration

The Ravel Runtime is responsible of all the VMs instances running on one host. It includes the Ravel Machines VMs and local VMs.

```toml
[daemon.runtime]
cloud_hypervisor_binary = "/opt/ravel/cloud-hypervisor" # Must be statically linked 
jailer_binary = "/opt/ravel/jailer" # Path to the Ravel jailer
init_binary = "./opt/ravel/initd"  # Path to Initd binary
linux_kernel = "/opt/ravel/vmlinux.bin" # A build of the cloud-hypervisor linux kernel
```

### Agent configuration

The Ravel Agent is responsible of managing workloads assigned to one host in the Ravel cluster.


```toml
[daemon.agent]
resources = { cpus_mhz = 20000, memory_mb = 16_384 } # Resources allocatable by the agent
node_id = "ravel-1" # MUST be unique in a cluster because its uniqueness is not enforced.
region = "fr" # The region the agent will announce itself in
address = "127.0.0.1" # The cluster wide reachable address of the agent
port = 8080 # The HTTP port the agent will listen on for the internal API
http_proxy_port = 8082 # The http proxy port the agent will announce as available to reach local VMs
```

In production environments, you definitly want to enable the mtls:
Theses certificates can be generated with the `ravel tls` commands.

```toml
[daemon.agent.tls]
cert_file = "ravel-1-agent-cert.pem"
key_file = "ravel-1-agent-key.pem"
ca_file = "ravel-ca-cert.pem"
```


## Server configuration

The Ravel server is responsible to accept API requests to schedule workloads on the cluster. The Ravel server store his state in a Postgres database and use HTTP and NATS to communicate with the agents.


```toml
[server]
postgres_url = "postgres://user:pass@host:5432/ravel-db" 
[server.api]
address = ":3000" # The HTTP address the server will listen on

# [server.api.tls]
# cert_file = ""
# key_file = ""
# ca_file = ""
[server.tls] # The server tls configuration used to communicate with the agents
cert_file = "server-1-server-cert.pem"
key_file = "server-1-server-key.pem"
ca_file = "ravel-ca-cert.pem"
```

Once the server is configure you can start it with the `ravel server` command.


### Machine templates

The ravel server must be configured with some machine templates to start accepting workloads.

The `vcpu_frequency` is the frequency allocated by virtual-cpus in MHz. This value is converted CPU quotas used by Ravel with cgroups to control the CPU usage of the VMs.
For each count of vcpus, you can define a list of memory configurations in MB.

```toml
[server.machine_templates.std]
vcpu_frequency = 2500
combinations = [
    { vcpus = 1, memory_configs = [
        1_024,
        2_048,
        4_096,
    ] },
    { vcpus = 2, memory_configs = [
        2_048,
        4_096,
        8_192,
    ] },
    { vcpus = 4, memory_configs = [
        4_096,
        8_192,
        16_384,
    ] },
]
```