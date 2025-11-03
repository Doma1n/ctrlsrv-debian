# ctrlsrv - Home Control Server

A Debian-based touchscreen control panel for home printing, file storage, and remote access. Built as a complete homelab infrastructure project showcasing systems programming, Go development, and modern networking protocols.

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go Version](https://img.shields.io/badge/go-1.24-blue.svg)
![Platform](https://img.shields.io/badge/platform-Debian%20Bookworm-red.svg)

## 🎯 Overview

This project transforms a Dell Inspiron 11 (touchscreen) into a home control server that provides:

- 📄 **Network Printing**: CUPS server with automated print dropbox
- 📁 **File Sharing**: Samba shares with storage monitoring
- 📱 **Touch UI**: Lightweight Go + Fyne interface optimized for embedded use
- 🔐 **Remote Access**: WireGuard VPN + xRDP for secure remote management
- 🌐 **Cloud Edge**: HTTP/3 (QUIC) proxy for low-latency external access
- 📊 **System Monitoring**: SMART disk monitoring, service health checks

## 🏗️ Architecture

```
┌─────────────┐      HTTP/3 (QUIC)       ┌──────────────┐
│   Mobile    │────────────────────────▶ │  AWS Edge   │
│   Client    │                          │  (Go Proxy)  │
└─────────────┘                          └──────┬───────┘
                                                │
                                          WireGuard VPN
                                         HTTP/3 (QUIC)
                                                │
                                         ┌──────▼───────┐
                                         │   ctrlsrv    │
                                         │  (Debian)    │
                                         └──────┬───────┘
                                                │
                        ┌───────────────────────┼────────────────┐
                        │                       │                │
                    ┌───▼────┐            ┌────▼─────┐    ┌─────▼──────┐
                    │  CUPS  │            │  Samba   │    │   Docker   │
                    │ Print  │            │  Shares  │    │ Containers │
                    └────────┘            └──────────┘    └────────────┘
                        │                       │
                   /srv/storage1 (mandatory external disk)
```

## ✨ Key Features

### Storage-First Design
- External disk `/srv/storage1` is **mandatory** - system won't boot without it
- All services depend on storage availability
- Runtime monitoring with automatic service shutdown if disk disappears
- Smart service dependency management via systemd

### Modern Networking
- **HTTP/3 (QUIC)** end-to-end - no protocol translation
- WireGuard VPN for secure remote access
- mTLS authentication for cloud edge
- Low-latency design optimized for mobile clients

### Touch-Optimized UI
- Go + Fyne UI (50-80MB memory vs 500MB+ for Chromium)
- Keyboard shortcuts for physical keyboard access
- Auto-restart on crashes
- F11 to exit fullscreen, multiple escape hatches

### Automated Deployment
- Custom Debian ISO with live-build
- Single-script installation (`setup-srv.sh`)
- All services pre-configured
- <30 minute rebuild time

## 🚀 Quick Start

### Prerequisites
- Dell Inspiron 11 (or similar, i3+, 4GB+ RAM, touchscreen optional)
- External USB disk for `/srv/storage1`
- Debian Bookworm (or use custom ISO)

### Installation

1. **Clone this repository**
```bash
git clone https://github.com/yourusername/ctrlsrv.git
cd ctrlsrv
```

2. **Build the custom ISO** (optional)
```bash
cd debian-iso
sudo apt install live-build
./build-iso.sh
# Write resulting ISO to USB with Rufus (Windows) or dd (Linux)
```

3. **Or install on existing Debian**
```bash
# Copy config.example.yaml to config.yaml and edit
cp config.example.yaml config.yaml
nano config.yaml

# Run setup script
sudo ./scripts/setup-srv.sh
```

4. **Build and install the Go daemon**
```bash
make build
sudo make install
```

5. **Configure storage**
```bash
# Find your external disk UUID
sudo blkid

# Add to /etc/fstab (NO nofail!)
echo "UUID=xxxx-xxxx /srv/storage1 ext4 defaults 0 2" | sudo tee -a /etc/fstab

# Mount and reboot
sudo systemctl daemon-reload
sudo mount -a
sudo reboot
```

## 📦 Components

### Core Daemon (`cmd/ctrlsrvd`)
Go application that provides:
- HTTP/1.1 + H2 API on `:8080` (local)
- HTTP/3 (QUIC) on `:8443` (edge/mobile)
- Fyne touch UI
- System service orchestration

### Services
- **CUPS**: Network printer (Canon TR4550)
- **SANE**: Network scanner
- **Samba**: File shares (`/srv/storage1`, `/srv/storage1/printdrop`)
- **Print Watcher**: Auto-prints PDFs dropped in printdrop
- **Storage Monitor**: Watches disk availability
- **xRDP**: Remote desktop access

### Scripts
- `setup-srv.sh`: Complete system setup
- `build-iso.sh`: Generate custom Debian ISO
- `storage-watch.sh`: Runtime storage monitoring

## 🔧 Configuration

### Example config.yaml
```yaml
server:
  listen_addr: "0.0.0.0:8080"
  quic_addr: "0.0.0.0:8443"
  
storage:
  path: "/srv/storage1"
  
cups:
  url: "http://localhost:631"
  printer: "Canon_TR4550"
  
edge:
  endpoint: "edge.example.com:443"
  
wireguard:
  interface: "wg0"
  allowed_networks:
    - "192.168.1.0/24"
    - "10.8.0.0/24"
```

See `config.example.yaml` for full configuration options.

## 🛠️ Development

### Build
```bash
make build          # Build binary
make test           # Run tests
make lint           # Run linters
make fmt            # Format code
```

### Run Locally
```bash
make run            # Run with example config
```

### Create ISO
```bash
make iso            # Build custom Debian ISO
```

## 📊 Tech Stack

- **Language**: Go 1.24
- **UI Framework**: Fyne (Go)
- **OS**: Debian Bookworm
- **Protocols**: HTTP/3 (QUIC), gRPC
- **Services**: CUPS, Samba, xRDP, WireGuard
- **Init System**: systemd
- **Build Tools**: live-build, make

## 🔒 Security

This project implements:
- mTLS for client authentication
- WireGuard VPN for network security
- Service isolation via systemd
- Principle of least privilege (sudoers restrictions)

**Important Security Notes:**
- Never expose CUPS/Samba/xRDP directly to the internet
- Use WireGuard or SSH tunnels for remote access
- Review and harden `setup-srv.sh` before production use
- Change all default passwords immediately

## 📝 Portfolio Note

This is a personal homelab project built for learning and demonstration purposes. It showcases:

- **Systems Programming**: Custom OS image creation, service orchestration
- **Go Development**: Microservices, QUIC implementation, concurrent design
- **Networking**: HTTP/3, WireGuard VPN, edge proxy patterns
- **DevOps**: Infrastructure as code, systemd, automation
- **UI Design**: Touch-optimized embedded interfaces

**For Recruiters/Employers**: Feel free to clone and test this project locally. If you have questions about implementation details, I'm happy to discuss!

**For Fellow Developers**: You're welcome to learn from this code. If you build something inspired by this, I'd appreciate attribution. PRs and issues are welcome!

## ⚠️ Disclaimer

This is a personal homelab project provided as-is for educational and portfolio purposes. It includes system-level configurations and should not be deployed in production environments without thorough security review and hardening.

The author is not responsible for:
- Data loss from storage failures
- Security vulnerabilities in home deployments
- Hardware damage from misconfiguration
- Any other issues arising from use of this software

**Use at your own risk. Always test in a safe environment first.**

## 🤝 Related Projects

- [homelab-edge](https://github.com/yourusername/homelab-edge) - AWS edge proxy (Go)
- [homelab-android](https://github.com/yourusername/homelab-android) - Android client (Kotlin)

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🔗 Links

- **Documentation**: [docs/](docs/)
- **Issue Tracker**: [GitHub Issues](https://github.com/yourusername/ctrlsrv/issues)
- **Author**: Doma1n

---

**Built with ❤️ for homelabs everywhere**