#!/bin/bash
# Setup Script for ctrlsrv - Home Control Server
# Run this after installing Debian or from custom ISO

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_step() { echo -e "${BLUE}[STEP]${NC} $1"; }

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    log_error "Please run as root (sudo ./setup-srv.sh)"
    exit 1
fi

echo "=========================================="
echo "  ctrlsrv Setup Script"
echo "=========================================="
echo ""

# 1. Create user ctrlsrv
log_step "Creating user ctrlsrv..."
if ! id ctrlsrv &>/dev/null; then
    useradd -m -s /bin/bash ctrlsrv
    echo "ctrlsrv:ctrlsrv" | chpasswd
    log_info "User ctrlsrv created (password: ctrlsrv - CHANGE THIS!)"
    log_warn "Run: passwd ctrlsrv to change password"
else
    log_warn "User ctrlsrv already exists"
fi

# Add to groups
usermod -aG sudo,lpadmin,sambashare ctrlsrv 2>/dev/null || true

# 2. Sudoers configuration
log_step "Configuring sudoers..."
cat > /etc/sudoers.d/ctrlsrv << 'EOF'
ctrlsrv ALL=(ALL) NOPASSWD: /usr/sbin/service, /bin/systemctl, /usr/bin/lp, /usr/sbin/smartctl, /usr/bin/systemctl
EOF
chmod 440 /etc/sudoers.d/ctrlsrv
log_info "Sudoers configured"

# 3. Create storage directory
log_step "Creating storage mountpoint..."
mkdir -p /srv/storage1
mkdir -p /srv/storage1/printdrop
chmod 777 /srv/storage1/printdrop
log_info "Storage directories created"

# 4. Configure CUPS
log_step "Configuring CUPS..."
if [ -f /etc/cups/cupsd.conf ]; then
    cp /etc/cups/cupsd.conf /etc/cups/cupsd.conf.bak

    cat > /etc/cups/cupsd.conf << 'EOF'
LogLevel warn
MaxLogSize 0
Listen 631
Listen /run/cups/cups.sock
Browsing On
BrowseLocalProtocols dnssd
DefaultAuthType Basic
WebInterface Yes

<Location />
  Order allow,deny
  Allow from 127.0.0.1
  Allow from 192.168.1.0/24
  Allow from 10.8.0.0/24
</Location>

<Location /admin>
  Order allow,deny
  Allow from 127.0.0.1
  Allow from 192.168.1.0/24
  Allow from 10.8.0.0/24
</Location>

<Location /admin/conf>
  AuthType Default
  Require user @SYSTEM
  Order allow,deny
  Allow from 127.0.0.1
  Allow from 192.168.1.0/24
  Allow from 10.8.0.0/24
</Location>
EOF

    systemctl enable cups
    systemctl restart cups
    log_info "CUPS configured for LAN + WireGuard access"
else
    log_warn "CUPS not installed, skipping configuration"
fi

# 5. Configure SANE
log_step "Configuring SANE..."
if [ -f /etc/sane.d/saned.conf ]; then
    cat > /etc/sane.d/saned.conf << 'EOF'
# Allow from LAN and WireGuard
192.168.1.0/24
10.8.0.0/24
EOF

    cat > /etc/default/saned << 'EOF'
RUN=yes
EOF

    systemctl enable saned.socket 2>/dev/null || true
    log_info "SANE configured"
else
    log_warn "SANE not installed, skipping configuration"
fi

# 6. Configure Samba
log_step "Configuring Samba..."
if [ -f /etc/samba/smb.conf ]; then
    cp /etc/samba/smb.conf /etc/samba/smb.conf.bak

    cat > /etc/samba/smb.conf << 'EOF'
[global]
   workgroup = WORKGROUP
   server string = %h server (Samba, Debian)
   log file = /var/log/samba/log.%m
   max log size = 1000
   logging = file
   panic action = /usr/share/samba/panic-action %d
   server role = standalone server
   obey pam restrictions = yes
   unix password sync = yes
   passwd program = /usr/bin/passwd %u
   passwd chat = *Enter\snew\s*\spassword:* %n\n *Retype\snew\s*\spassword:* %n\n *password\supdated\ssuccessfully* .
   pam password change = yes
   map to guest = bad user
   usershare allow guests = yes

[storage]
   comment = Main Storage
   path = /srv/storage1
   browseable = yes
   read only = no
   guest ok = yes
   create mask = 0644
   directory mask = 0755

[printdrop]
   comment = Print Drop Folder
   path = /srv/storage1/printdrop
   browseable = yes
   read only = no
   guest ok = yes
   create mask = 0666
   directory mask = 0777
EOF

    systemctl enable smbd nmbd
    systemctl restart smbd nmbd
    log_info "Samba configured"
else
    log_warn "Samba not installed, skipping configuration"
fi

# 7. Install Docker
log_step "Installing Docker..."
if ! command -v docker &>/dev/null; then
    curl -fsSL https://get.docker.com -o /tmp/get-docker.sh
    sh /tmp/get-docker.sh
    usermod -aG docker ctrlsrv

    # Configure Docker data root
    mkdir -p /srv/storage1/docker
    mkdir -p /etc/docker
    cat > /etc/docker/daemon.json << 'EOF'
{
  "data-root": "/srv/storage1/docker"
}
EOF
    systemctl enable docker
    systemctl restart docker
    log_info "Docker installed with data-root on /srv/storage1"
else
    log_warn "Docker already installed"
fi

# 8. Configure LightDM autologin
log_step "Configuring LightDM autologin..."
mkdir -p /etc/lightdm/lightdm.conf.d
cat > /etc/lightdm/lightdm.conf.d/50-autologin.conf << 'EOF'
[Seat:*]
autologin-user=ctrlsrv
autologin-user-timeout=0
EOF
log_info "LightDM configured for autologin"

# 9. Create Openbox autostart
log_step "Creating Openbox autostart..."
mkdir -p /home/ctrlsrv/.config/openbox
cat > /home/ctrlsrv/.config/openbox/autostart << 'EOF'
#!/bin/bash

# Disable screen blanking
xset -dpms
xset s off

# Hide cursor after 1 second idle
unclutter -idle 1 &

# On-screen keyboard
onboard &

# VMware tools (if in VM)
if command -v vmware-user &>/dev/null; then
    vmware-user &
fi

# Wait for ctrlsrvd to be ready
sleep 3

# Open kiosk browser (surf preferred)
if command -v surf &>/dev/null; then
    surf -F http://localhost:8080 &
elif command -v netsurf-gtk &>/dev/null; then
    netsurf-gtk -f http://localhost:8080 &
else
    chromium --kiosk --app=http://localhost:8080 --incognito &
fi

# Watchdog: restart browser if it crashes
while true; do
    sleep 5
    if ! pgrep -x surf &>/dev/null && \
       ! pgrep -x netsurf-gtk &>/dev/null && \
       ! pgrep -x chromium &>/dev/null; then
        logger -t openbox "Browser crashed, restarting..."
        surf -F http://localhost:8080 &
    fi
done
EOF

chmod +x /home/ctrlsrv/.config/openbox/autostart
chown -R ctrlsrv:ctrlsrv /home/ctrlsrv/.config
log_info "Openbox autostart configured"

# 10. Enable xRDP
log_step "Configuring xRDP..."
if command -v xrdp &>/dev/null; then
    systemctl enable xrdp
    systemctl restart xrdp
    log_info "xRDP enabled (bind to WireGuard after configuration)"
else
    log_warn "xRDP not installed, skipping configuration"
fi

# 11. Enable smartd
log_step "Enabling smartd..."
if command -v smartd &>/dev/null; then
    systemctl enable smartd
    systemctl restart smartd
    log_info "smartd enabled"
else
    log_warn "smartd not installed, skipping configuration"
fi

# 12. Install print-watcher
log_step "Installing print-watcher..."
if [ -f scripts/print-watcher.sh ]; then
    cp scripts/print-watcher.sh /usr/local/bin/
    chmod +x /usr/local/bin/print-watcher.sh

    if [ -f config/systemd/print-watcher.service ]; then
        cp config/systemd/print-watcher.service /etc/systemd/system/
        systemctl daemon-reload
        systemctl enable print-watcher
        log_info "Print watcher installed"
    fi
else
    log_warn "print-watcher.sh not found, skipping"
fi

# 13. Install storage-watch
log_step "Installing storage-watch..."
if [ -f scripts/storage-watch.sh ]; then
    cp scripts/storage-watch.sh /usr/local/bin/
    chmod +x /usr/local/bin/storage-watch.sh

    if [ -f config/systemd/storage-watch.service ]; then
        cp config/systemd/storage-watch.service /etc/systemd/system/
        systemctl daemon-reload
        systemctl enable storage-watch
        log_info "Storage watch installed"
    fi
else
    log_warn "storage-watch.sh not found, skipping"
fi

# 14. Set hostname
log_step "Setting hostname..."
echo "ctrlsrv" > /etc/hostname
cat > /etc/hosts << 'EOF'
127.0.0.1       localhost
127.0.1.1       ctrlsrv.local   ctrlsrv

# The following lines are desirable for IPv6 capable hosts
::1     localhost ip6-localhost ip6-loopback
ff02::1 ip6-allnodes
ff02::2 ip6-allrouters
EOF
log_info "Hostname set to ctrlsrv"

# 15. Enable Avahi
log_step "Enabling Avahi..."
if command -v avahi-daemon &>/dev/null; then
    systemctl enable avahi-daemon
    systemctl restart avahi-daemon
    log_info "Avahi enabled (ctrlsrv.local)"
else
    log_warn "Avahi not installed, skipping"
fi

# 16. System configuration
log_step "Configuring system..."

# Enable kernel panic auto-reboot
sysctl -w kernel.panic=10 2>/dev/null || true
echo "kernel.panic = 10" > /etc/sysctl.d/99-panic.conf

log_info "System configuration complete"

# Summary
echo ""
echo "=========================================="
log_info "Setup Complete!"
echo "=========================================="
echo ""
log_warn "IMPORTANT NEXT STEPS:"
echo "1. Change ctrlsrv password: passwd ctrlsrv"
echo "2. Edit /etc/fstab to add /srv/storage1 (UUID based, no nofail):"
echo "   sudo blkid  # Find UUID"
echo "   echo 'UUID=xxxx-xxxx /srv/storage1 ext4 defaults 0 2' | sudo tee -a /etc/fstab"
echo "3. Build and install ctrlsrvd:"
echo "   cd /path/to/ctrlsrv"
echo "   make build"
echo "   sudo make install"
echo "4. Configure WireGuard in /etc/wireguard/wg0.conf"
echo "5. Edit /etc/xrdp/xrdp.ini to bind to 10.8.0.2 (WireGuard IP)"
echo "6. Reboot: sudo reboot"
echo ""
log_info "Access CUPS at: http://ctrlsrv.local:631"
log_info "Access Samba shares at: \\\\ctrlsrv.local\\storage"
echo ""