# Ravel thinpool
## Containerd thinpool

Ravel use containerd configured with devicemapper storage driver and thinpool to build virtual machines rootfs. You need to provide to containerd the thinpool configuration.


See the [docker documentation for production setup](https://github.com/docker/docs/blob/4b74397d7ae2bfb0d32fa4aa0aa3847af2055291/content/storage/storagedriver/device-mapper-driver.md#configure-direct-lvm-mode-manually) for more information about the thinpool configuration in production.


### For development

For development you can follow the production doc just by creating a loopback device to use for the thinpool.

**1. Create a backing file for the loopback device**
```
$ fallocate -l 30G ./tmp/vg-ravel.img
```

**2. Create a loopback device**
Find an available loopback device
```
$ losetup -f
```
Exemple output:

```
/dev/loop10
```

Create the loopback device
```
$ losetup /dev/loop0 ./tmp/vg-ravel.img
```

Then you can follow the production doc to configure the thinpool with the loopback device by replacing "/dev/xvdf" with the loopback device you just created.