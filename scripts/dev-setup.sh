#!/bin/bash

# Ravel Development Environment Setup Script
# This script sets up a complete development environment for Ravel

set -e

COLOR_RED='\033[0;31m'
COLOR_GREEN='\033[0;32m'
COLOR_YELLOW='\033[1;33m'
COLOR_BLUE='\033[0;34m'
COLOR_RESET='\033[0m'

log_info() {
    echo -e "${COLOR_BLUE}[INFO]${COLOR_RESET} $1"
}

log_success() {
    echo -e "${COLOR_GREEN}[SUCCESS]${COLOR_RESET} $1"
}

log_warning() {
    echo -e "${COLOR_YELLOW}[WARNING]${COLOR_RESET} $1"
}

log_error() {
    echo -e "${COLOR_RED}[ERROR]${COLOR_RESET} $1"
}

check_command() {
    if command -v $1 &> /dev/null; then
        log_success "$1 is installed"
        return 0
    else
        log_warning "$1 is not installed"
        return 1
    fi
}

echo "======================================"
echo "  Ravel Development Setup"
echo "======================================"
echo ""

# Check if running with sudo
if [ "$EUID" -ne 0 ]; then
    log_warning "Some operations require sudo access"
    USE_SUDO="sudo"
else
    USE_SUDO=""
fi

# 1. Check Go installation
log_info "Checking Go installation..."
if check_command go; then
    GO_VERSION=$(go version | awk '{print $3}')
    log_info "Go version: $GO_VERSION"

    # Check if version is >= 1.22
    REQUIRED_VERSION="1.22"
    CURRENT_VERSION=$(go version | sed 's/go version go//' | awk '{print $1}')

    if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$CURRENT_VERSION" | sort -V | head -n1)" = "$REQUIRED_VERSION" ]; then
        log_success "Go version is sufficient"
    else
        log_error "Go version $CURRENT_VERSION is too old. Please install Go 1.22 or newer"
        exit 1
    fi
else
    log_error "Go is not installed. Please install Go 1.22 or newer from https://golang.org/dl/"
    exit 1
fi

# 2. Check KVM support
log_info "Checking KVM support..."
if lsmod | grep -q kvm; then
    log_success "KVM module is loaded"
else
    log_error "KVM is not enabled. Please enable KVM in your BIOS and load the module"
    exit 1
fi

# 3. Check TUN/TAP support
log_info "Checking TUN/TAP support..."
if [ -e /dev/net/tun ]; then
    log_success "TUN/TAP is available"
else
    log_warning "TUN/TAP may not be available. Creating /dev/net/tun..."
    $USE_SUDO mkdir -p /dev/net
    $USE_SUDO mknod /dev/net/tun c 10 200 2>/dev/null || true
    $USE_SUDO chmod 0666 /dev/net/tun
fi

# 4. Install development dependencies
log_info "Installing development dependencies..."

# Install golangci-lint
if ! check_command golangci-lint; then
    log_info "Installing golangci-lint..."
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    log_success "golangci-lint installed"
fi

# 5. Download Go module dependencies
log_info "Downloading Go module dependencies..."
go mod download
log_success "Go modules downloaded"

# 6. Create required directories
log_info "Creating required directories..."
$USE_SUDO mkdir -p /var/lib/ravel
$USE_SUDO mkdir -p /opt/ravel
$USE_SUDO mkdir -p /etc/ravel
mkdir -p bin
log_success "Directories created"

# 7. Check for Cloud Hypervisor
log_info "Checking for Cloud Hypervisor..."
if [ -f /opt/ravel/cloud-hypervisor ]; then
    log_success "Cloud Hypervisor found at /opt/ravel/cloud-hypervisor"
else
    log_warning "Cloud Hypervisor not found"
    read -p "Do you want to download Cloud Hypervisor? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        log_info "Downloading Cloud Hypervisor..."
        wget -q https://github.com/cloud-hypervisor/cloud-hypervisor/releases/download/v38.0/cloud-hypervisor-static -O /tmp/cloud-hypervisor
        $USE_SUDO mv /tmp/cloud-hypervisor /opt/ravel/cloud-hypervisor
        $USE_SUDO chmod +x /opt/ravel/cloud-hypervisor
        log_success "Cloud Hypervisor installed"

        log_info "Downloading Linux kernel..."
        wget -q https://github.com/cloud-hypervisor/cloud-hypervisor/releases/download/v38.0/vmlinux -O /tmp/vmlinux
        $USE_SUDO mv /tmp/vmlinux /opt/ravel/vmlinux.bin
        log_success "Linux kernel installed"
    fi
fi

# 8. Check for containerd
log_info "Checking for containerd..."
if check_command containerd; then
    if systemctl is-active --quiet containerd; then
        log_success "containerd is running"
    else
        log_warning "containerd is installed but not running"
        read -p "Do you want to start containerd? (y/n) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            $USE_SUDO systemctl start containerd
            $USE_SUDO systemctl enable containerd
            log_success "containerd started"
        fi
    fi
else
    log_warning "containerd is not installed"
    echo "Please install containerd:"
    echo "  Ubuntu/Debian: sudo apt-get install containerd"
    echo "  Fedora/RHEL: sudo dnf install containerd"
fi

# 9. Build Ravel binaries
log_info "Building Ravel binaries..."
make build-all
log_success "Ravel binaries built"

# 10. Check for PostgreSQL (optional)
log_info "Checking for PostgreSQL (optional for clustering)..."
if check_command psql; then
    log_success "PostgreSQL client is installed"
else
    log_warning "PostgreSQL client not found (optional - needed for clustering)"
fi

# 11. Check for NATS (optional)
log_info "Checking for NATS (optional for clustering)..."
if check_command nats-server; then
    log_success "NATS server is installed"
else
    log_warning "NATS server not found (optional - needed for clustering)"
fi

# 12. Create sample configuration
log_info "Creating sample configuration..."
if [ ! -f ravel.toml ]; then
    cat > ravel.toml <<EOF
[daemon]
database_path = "/var/lib/ravel/daemon.db"

[daemon.runtime]
cloud_hypervisor_binary = "/opt/ravel/cloud-hypervisor"
jailer_binary = "/opt/ravel/jailer"
init_binary = "/opt/ravel/initd"
linux_kernel = "/opt/ravel/vmlinux.bin"

[daemon.agent]
resources = { cpus_mhz = 8000, memory_mb = 8192 }
node_id = "dev-node-1"
region = "local"
address = "127.0.0.1"
port = 8080
EOF
    log_success "Sample configuration created: ravel.toml"
else
    log_info "Configuration file already exists: ravel.toml"
fi

# 13. Summary
echo ""
echo "======================================"
log_success "Development environment setup complete!"
echo "======================================"
echo ""
echo "Next steps:"
echo "  1. Review and edit ravel.toml configuration"
echo "  2. Build binaries: make build-all"
echo "  3. Run tests: make test"
echo "  4. Start daemon: make run-daemon"
echo ""
echo "For more information, see docs/quickstart.md"
echo ""
