#!/bin/bash
set -e

thin_pool_autoextend_threshold=80
thin_pool_autoextend_percent=20
file_path=$1
pool_name=$2

# Find an unused loop device
loop_device=$(losetup -f) || {
    echo "Error: No unused loop devices available."
    exit 1
}

losetup $loop_device $file_path

echo "Loop device created: $loop_device"

# Create a thin pool
echo "Creating thin pool $pool_name"

echo "Creating physical volume and volume group $pool_name"
pvcreate $loop_device
vgcreate $pool_name $loop_device

echo "Creating logical volumes"
lvcreate --wipesignatures y -n thinpool $pool_name -l 95%VG
lvcreate --wipesignatures y -n thinpoolmeta $pool_name -l 1%VG
echo "Converting to thin pool"

# Convert the volumes to a thin pool and a storage location for metadata for the thin pool, using the lvconvert command.
sudo lvconvert -y \
    --zero n \
    -c 512K \
    --thinpool $pool_name/thinpool \
    --poolmetadata $pool_name/thinpoolmeta

cat >/etc/lvm/profile/$pool_name-thinpool.profile <<EOF
activation {
    thin_pool_autoextend_threshold=$thin_pool_autoextend_threshold
    thin_pool_autoextend_percent=$thin_pool_autoextend_percent
}
EOF

lvchange --metadataprofile $pool_name-thinpool $pool_name/thinpool

# Activate the thin pool monitoring
lvchange --monitor y $pool_name/thinpool

cat <<EOF
#
# Add this to your contaienrd config.toml configuration file and restart the containerd daemon
#
[plugins]
  [plugins."io.containerd.snapshotter.v1.devmapper"]
    pool_name = "$pool_name"
    root_path = "/var/lib/containerd/devmapper"
    base_image_size = "8GB"
    discard_blocks = true
EOF
