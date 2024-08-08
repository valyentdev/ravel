#!/bin/bash

set -e

create_loop_drive() {
    size=$1
    name=$2

    # Generate a random name if not provided
    if [ -z "$name" ]; then
        name="ravel-$(date +%s%N | sha256sum | head -c8)"
    fi

    mkdir -p ./disks
    file_path="./disks/$name"

    # Create a file of a certain size
    fallocate -l $size $file_path

    # Find the first unused loop device
    loop_device=$(sudo losetup -f) || {
        echo "Error: No unused loop devices available."
        exit 1
    }

    # Set up a loop device for the file
    sudo losetup $loop_device $file_path

    # Output the created name
    echo "$name $loop_device"
}

create_volume_group() {
    size=$1
    name=$2

    # Create loop drive
    read drive_name loop_device <<< $(create_loop_drive $size $name)

    # Create a new volume group with the loop device
    echo "$drive_name $loop_device"
    
    if [ -z "$loop_device" ]; then
        echo "error loop device empty"
        return 5
    fi

    sudo vgcreate $drive_name $loop_device
}

list_volume_groups() {
    echo "  VG             #PV #LV #SN Attr   VSize VFree"
    # Run vgs and grep for volume group names containing "ravel"
    vgs_output=$(sudo vgs | grep 'ravel')

    # Print the selected result
    echo "$vgs_output"
}

get_loop_device() {
    volume_group=$1

    # Run pvs and parse its output
    while read -r line; do
        # Split the line into words
        read -a words <<< "$line"

        # Check if the volume group name matches
        if [[ "${words[1]}" == "$volume_group" ]]; then
            # If it matches, output the loop device and return
            echo "${words[0]}"
            return
        fi
    done <<< "$(sudo pvs)"
}

delete_all_logical_volumes() {
    volume_group=$1

    # Run lvs and parse its output
    while read -r lv vg; do
        # Check if the volume group name matches
        if [[ "$vg" == "$volume_group" ]]; then
            # If it matches, remove the logical volume
            sudo lvremove -y "$vg/$lv"
        fi
    done <<< "$(sudo lvs --noheadings -o lv_name,vg_name)"

    echo "All logical volumes in volume group $volume_group have been deleted."
}

delete_volume_group() {
    volume_group=$1
    
    echo "Deleting All logical volumes"
    # Remove logical volumes
    delete_all_logical_volumes "$volume_group"

    loop_device=$(get_loop_device "$volume_group")
    if [ -z "$loop_device" ]; then
        echo "couldn't find loop device"
        return 5
    fi

    # Remove volume group
    echo "removing volume group $volume_group"
    sudo vgremove $volume_group

    # Remove loop device
    echo "deleting lop device $loop_device"
    sudo losetup -d $loop_device

    # Remove the file
    echo "removing backing file"
    sudo rm -f ./disks/$volume_group
}

# Main script logic
case $1 in
    "create")
        create_volume_group $2 $3
        ;;
    "list")
        list_volume_groups
        ;;
    "delete")
        delete_volume_group $2
        ;;
    *)
        echo "Usage: $0 {create <size> [name] | list | delete <volume-group>}"
        exit 1
        ;;
esac
