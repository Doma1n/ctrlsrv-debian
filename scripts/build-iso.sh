#!/bin/bash
# Build custom Debian ISO with ctrlsrv pre-configured

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    log_error "Please run as root (sudo ./build-iso.sh)"
    exit 1
fi

# Check dependencies
log_info "Checking dependencies..."
for cmd in lb debootstrap xorriso; do
    if ! command -v $cmd &>/dev/null; then
        log_error "$cmd is not installed"
        echo "Install with: sudo apt install live-build debootstrap xorriso isolinux syslinux-efi"
        exit 1
    fi
done

log_info "Starting custom Debian ISO build..."
echo "This will take 10-30 minutes depending on your internet connection."
echo ""

# Clean previous build
if [ -d config ] || [ -d binary ] || [ -d chroot ]; then
    log_warn "Cleaning previous build..."
    lb clean --purge
fi

# Configure live-build
log_info "Configuring live-build..."
lb config \
    --architectures amd64 \
    --distribution bookworm \
    --debian-installer live \
    --debian-installer-gui false \
    --archive-areas "main contrib non-free non-free-firmware" \
    --bootappend-live "boot=live components quiet splash" \
    --iso-application "ctrlsrv Control Server" \
    --iso-volume "CTRLSRV" \
    --memtest none

# Create package lists directory
mkdir -p config/package-lists

# Create package list
log_info "Creating package list..."
cat > config/package-lists/ctrlsrv.list.chroot << 'EOF'
# Base system
openssh-server
vim
htop
curl
wget
git
avahi-daemon
chrony
sudo
tmux
rsync
net-tools

# Printing & Scanning
cups
printer-driver-all
printer-driver-gutenprint
hplip
system-config-printer
sane-utils
sane-airscan

# File Sharing
samba
cifs-utils

# Monitoring
smartmontools
lm-sensors

# File watching
inotify-tools

# GUI & Kiosk
xorg
lightdm
openbox
obconf
x11-xserver-utils
xinput
xdotool
unclutter
onboard
xfce4-terminal
pcmanfm
lxappearance
surf

# Remote Access
xrdp
tightvncserver

# Networking
wireguard
wireguard-tools
iptables
dnsutils

# Development
build-essential
EOF

log_info "Package list created"

# Copy setup script into the image
log_info "Adding setup script to image..."
mkdir -p config/includes.chroot/root
if [ -f ../scripts/setup-srv.sh ]; then
    cp ../scripts/setup-srv.sh config/includes.chroot/root/
    chmod +x config/includes.chroot/root/setup-srv.sh
else
    log_warn "setup-srv.sh not found, creating placeholder"
    echo "#!/bin/bash" > config/includes.chroot/root/setup-srv.sh
    echo "echo 'Setup script placeholder'" >> config/includes.chroot/root/setup-srv.sh
    chmod +x config/includes.chroot/root/setup-srv.sh
fi

# Build the ISO
log_info "Building ISO (this will take a while)..."
lb build 2>&1 | tee build.log

# Check if build succeeded
if [ -f live-image-amd64.hybrid.iso ]; then
    # Rename to something more descriptive
    ISO_NAME="ctrlsrv-$(date +%Y%m%d).iso"
    mv live-image-amd64.hybrid.iso "../$ISO_NAME"

    log_info "Build complete!"
    log_info "ISO created: $ISO_NAME"
    log_info "Size: $(du -h ../$ISO_NAME | cut -f1)"
    echo ""
    echo "Write to USB with:"
    echo "  Rufus (Windows) - Use DD mode"
    echo "  sudo dd if=$ISO_NAME of=/dev/sdX bs=4M status=progress (Linux)"
    echo ""
else
    log_error "Build failed! Check build.log for details"
    exit 1
fi