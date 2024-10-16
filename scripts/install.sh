#!/bin/bash

set -e

latest() {
    curl -s https://api.github.com/repos/valyentdev/ravel/releases/latest | grep "tag_name" | cut -d '"' -f 4
}

installLatest() {
    version=$(latest)
    echo "Installing latest version $version"

    version_without_v=$(echo $version | sed 's/^v//')

    download_url="https://github.com/valyentdev/ravel/releases/download/$version/ravel_${version_without_v}_linux_amd64.tar.gz"
    echo "Downloading $download_url"

    tmp_dir=$(mktemp -d)

    curl -q --fail --location --progress-bar --output "$tmp_dir/ravel.tar.gz" "$download_url"
    tar -C $tmp_dir -xzf $tmp_dir/ravel.tar.gz
    chmod +x $tmp_dir/ravel

    if [ -f /usr/sbin/ravel ]; then
        echo "Removing old version"
        rm /usr/sbin/ravel
    fi
    mv $tmp_dir/ravel /usr/sbin/ravel
    rm -r $tmp_dir
    echo "Installed $version"
}

installLatest
