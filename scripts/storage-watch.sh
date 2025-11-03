#!/bin/bash
# Storage Watch - Monitors storage availability and stops services if missing

set -euo pipefail

# Configuration
STORAGE_PATH="${STORAGE_PATH:-/srv/storage1}"
CHECK_INTERVAL="${CHECK_INTERVAL:-10}"
SERVICES_TO_STOP=("print-watcher" "docker" "ctrlsrvd")

# Colors for logging
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
    logger -t storage-watch "$1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
    logger -t storage-watch -p user.warning "$1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
    logger -t storage-watch -p user.err "$1"
}

# Track storage state
storage_was_available=false

check_storage() {
    # Check if path exists
    if [ ! -d "$STORAGE_PATH" ]; then
        return 1
    fi

    # Check if it's a mountpoint
    if ! mountpoint -q "$STORAGE_PATH"; then
        return 1
    fi

    # Check if writable
    if ! [ -w "$STORAGE_PATH" ]; then
        return 1
    fi

    return 0
}

stop_dependent_services() {
    log_error "Storage unavailable - stopping dependent services"

    for service in "${SERVICES_TO_STOP[@]}"; do
        if systemctl is-active --quiet "$service"; then
            log_warn "Stopping service: $service"
            systemctl stop "$service" 2>/dev/null || true
        fi
    done

    # Send notification if gotify is available
    # TODO: Add gotify notification support
}

start_dependent_services() {
    log_info "Storage available - starting dependent services"

    for service in "${SERVICES_TO_STOP[@]}"; do
        if systemctl is-enabled --quiet "$service"; then
            log_info "Starting service: $service"
            systemctl start "$service" 2>/dev/null || true
        fi
    done
}

log_info "Storage monitor starting..."
log_info "Monitoring: $STORAGE_PATH"
log_info "Check interval: ${CHECK_INTERVAL}s"

while true; do
    if check_storage; then
        if [ "$storage_was_available" = false ]; then
            log_info "Storage is now available: $STORAGE_PATH"
            start_dependent_services
            storage_was_available=true
        fi
    else
        if [ "$storage_was_available" = true ]; then
            log_error "CRITICAL: Storage disappeared: $STORAGE_PATH"
            stop_dependent_services
            storage_was_available=false
        elif [ "$storage_was_available" = false ]; then
            # Still not available, log periodically
            log_warn "Storage still unavailable: $STORAGE_PATH"
        fi
    fi

    sleep "$CHECK_INTERVAL"
done