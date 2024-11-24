# Ravel and containerd

Ravel use containerd to manages images and to build virtual machines rootfs.
Ravel use containerd configured with devicemapper storage driver and thinpool to build virtual machines rootfs. You need to provide to containerd the thinpool configuration.


### For development

For development you can follow the production doc just by creating a loopback device to use for the thinpool.

**1. Create a backing file for the loopback device**
```
$ fallocate -l 30G ./tmp/vg-ravel.img
```

Then you can use this [script][./scripts/setup-dev-thinpool.sh] to automatically setup the thinpool.

Alternatively follow the documentation of [containerd](https://github.com/containerd/containerd/blob/main/docs/snapshotters/devmapper.md) and this [docker tutorial](https://github.com/docker/docs/blob/4b74397d7ae2bfb0d32fa4aa0aa3847af2055291/content/storage/storagedriver/device-mapper-driver.md#configure-direct-lvm-mode-manually).

