# Containerd Runtime Driver

This package provides a native containerd runtime driver for Ravel as an alternative to the Cloud Hypervisor microVM driver.

## Purpose

The containerd driver allows running workloads as regular containers instead of microVMs, providing:

- **Lower overhead**: No VM boot time, less memory overhead
- **Faster startup**: Containers start in milliseconds vs seconds for VMs
- **Resource efficiency**: Direct process execution without hypervisor layer
- **Mixed deployments**: Run both VMs and containers in the same cluster

## Trade-offs

**Containerd Driver (this):**
- ✓ Lower resource overhead
- ✓ Faster startup
- ✓ Better density (more containers per host)
- ✗ Less isolation (kernel shared with host)
- ✗ No hardware virtualization

**VM Driver (Cloud Hypervisor):**
- ✓ Strong isolation (full VM boundary)
- ✓ Hardware virtualization support
- ✓ Custom kernels
- ✗ Higher overhead
- ✗ Slower startup

## Current Status

**PARTIALLY IMPLEMENTED** - The driver structure is in place but needs completion:

### Implemented:
- ✓ Driver interface implementation
- ✓ Resource limits via cgroups (CPU MHz, memory)
- ✓ Snapshot management via containerd
- ✓ Instance lifecycle structure

### TODO:
- [ ] OCI runtime spec generation
- [ ] Container creation and task execution
- [ ] Signal handling and process management
- [ ] Exec support for running commands in containers
- [ ] Task recovery after daemon restart
- [ ] Network namespace setup
- [ ] Volume mount support

## Architecture

```
Instance Request
    ↓
Containerd Driver
    ↓
┌─────────────┬─────────────┬──────────────┐
│  Snapshot   │   Cgroup    │  Container   │
│  (overlayfs)│  (v2)       │   Task       │
└─────────────┴─────────────┴──────────────┘
    ↓             ↓              ↓
  Image        Resources      OCI Runtime
  Layers       (CPU/Mem)      (runc/crun)
```

## Usage

When fully implemented, the runtime will be configurable to use either driver:

```toml
[runtime]
driver = "containerd"  # or "vm" for Cloud Hypervisor
```

Machines can then be scheduled to run as containers or VMs based on their requirements.
