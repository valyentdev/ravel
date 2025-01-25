# The Ravel Runtime

The Ravel Runtime is the most low-level component of Ravel. It is responsible for running the OCI images provided by the Ravel Agent / Daemon inside cloud-hypervisors VMs.

The Ravel Runtime uses three main components to run the VMs:
- Containerd: to manage the OCI images as VM rootfs with the devmapper snapshotter.
- Cloud-hypervisor: to run the VMs.
- The ravel Jailer to further isolate the VMs inside cgoups, mount, pid and network namespaces
- Initd: injected as the init process of the VMs to run the user provided entrypoint.

## Getting started

### Prerequisites

> **Note:** For now Ravel only supports Linux amd64 architectures with support for KVM.

1. KVM

Before anything, you need to check that your system is KVM enabled. You can do this by running the following command:

```bash
$ lsmod | grep kvm # Check the presence of the kvm module
# or
$ kvm-ok # if installed
```

2. TUN/TAP device
You need to enable the TUN/TAP device driver on your system. You can follow instructions from the Linux Kernel documentation [here](https://docs.kernel.org/networking/tuntap.html#configuration):
```
mkdir /dev/net (if it doesn't exist already)
mknod /dev/net/tun c 10 200
```

3. Cloud-hypervisor
You can download the [Cloud-hypervisor v43.0 release](https://github.com/cloud-hypervisor/cloud-hypervisor/releases/tag/v43.0) on github. You MUST download the statically linked binary. Then you can make it available at ```/opt/ravel/cloud-hypervisor```

4. [Containerd](https://github.com/containerd/containerd) installed and configured to run with the `devmapper` snapshotter.


### Install Ravel

1. Download the latest release of Ravel from the [releases page](https://github.com/valyentdev/ravel/releases).

2. Extract the archive and move the binaries to your PATH:

```bash
mkdir -p /opt/ravel
mkdir -p /etc/ravel
pushd $(mktemp -d)
tar -xvf ravel ravel_0.7.2_linux_amd64.tar.gz
mv -t /usr/sbin/ ravel ravel-proxy
mv -t /opt/ravel jailer initd
popd
```

2. Testing the installation

```$ ravel
$ ravel
A cli tool for managing raveld.

Usage:
  ravel [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  corrosion   Commands to interact with the Corrosion instance
  daemon      Start the Ravel daemon
  db          Database management commands
  disks       Manage disks
  help        Help about any command
  image       Manage images
  instance    Manage ravel instances
  server      Start the API server
  tls         Ravel TLS certificates management for mTLS

Flags:
      --debug   Enable debug logging
  -h, --help    help for ravel

Use "ravel [command] --help" for more information about a command.
```


3. Download the cloud-hypervisor linux kernel. A pre-built version is ready to use [here](https://github.com/valyentdev/linux-cloud-hypervisor/releases/tag/5.18.8). Alternatively you can built it by following the [cloud-hypervisor documentation](https://www.cloudhypervisor.org/docs/prologue/quick-start/#building-your-kernel). Then, make the uncompressed file available at `/opt/ravel/vmlinux.bin`.

### Configuration

In `/etc/ravel/config.toml` you can configure the Ravel Agent. Here is an example configuration:

```toml
[daemon]
database_path = "/var/lib/ravel/agent.db"
[daemon.runtime]
init_binary = "/opt/ravel/initd"
jailer_binary = "/opt/ravel/jailer"
cloud_hypervisor_binary = "./cloud-hypervisor"
linux_kernel = "./vmlinux.bin"
```
To learn more about the configuration options, see the [configuration documentation](./config.md).


### Running the Ravel Daemon

To start the Ravel Daemon, run the following command:

```bash
sudo ravel daemon [-c /etc/ravel/config.toml]
```

Then you can try to create a new instance with the following command:

```bash
ravel instance create -c instance.json
```

You can find an instance configuration example in the [examples](./examples/instance.json) directory.

