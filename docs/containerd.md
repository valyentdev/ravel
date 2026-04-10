# Ravel and containerd

Ravel use containerd to manages images and to build virtual machines rootfs.
Ravel use containerd configured with devmapper storage driver and a thinpool to build virtual machines rootfs. You need to provide to containerd the thinpool configuration.


### For development

**1. Create a backing file for the loopback device**
```
$ fallocate -l 30G ./tmp/vg-ravel.img
```

**2. Run the script**

You can use this [script](./scripts/setup-dev-thinpool.sh) to automatically setup the thinpool like this:

```bash
$ ./scripts/setup-dev-thinpool.sh ./tmp/vg-ravel.img ravel
```

Alternatively follow the documentation of [containerd](https://github.com/containerd/containerd/blob/main/docs/snapshotters/devmapper.md).

### For production

Please follow the production documentation provided by docker [here](https://docs.docker.com/engine/storage/drivers/device-mapper-driver/#configure-direct-lvm-mode-for-production).




