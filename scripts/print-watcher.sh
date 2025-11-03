#!/bin/bash
# Print Watcher - Automatically prints PDFs dropped in printdrop folder

set -euo pipefail

# Configuration
WATCH_DIR="${WATCH_DIR:-/srv/storage1/printdrop}"
PRINTER="${PRINTER:-Canon_TR4550}"
MAX_FILE_SIZE=$((100 * 1024 * 1024))  # 100MB limit

# Colors for logging
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
    logger -t print-watcher "$1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
    logger -t print-watcher -p user.warning "$1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
    logger -t print-watcher -p user.err "$1"
}

# Check if watch directory exists
if [ ! -d "$WATCH_DIR" ]; then
    log_error "Watch directory not found: $WATCH_DIR"
    log_error "Storage may not be mounted!"
    exit 1
fi

# Check if printer exists
if ! lpstat -p "$PRINTER" &>/dev/null; then
    log_warn "Printer '$PRINTER' not found. Continuing anyway..."
fi

log_info "Print watcher starting..."
log_info "Watching: $WATCH_DIR"
log_info "Printer: $PRINTER"

# Function to print a file
print_file() {
    local file="$1"
    local filename=$(basename "$file")

    # Check if file still exists (might have been deleted)
    if [ ! -f "$file" ]; then
        log_warn "File disappeared: $filename"
        return
    fi

    # Check file size
    local size=$(stat -c%s "$file")
    if [ "$size" -gt "$MAX_FILE_SIZE" ]; then
        log_error "File too large (${size} bytes): $filename"
        # Move to error folder
        mkdir -p "$WATCH_DIR/errors"
        mv "$file" "$WATCH_DIR/errors/"
        return
    fi

    # Verify it's actually a PDF
    if ! file "$file" | grep -q "PDF"; then
        log_error "Not a valid PDF: $filename"
        mkdir -p "$WATCH_DIR/errors"
        mv "$file" "$WATCH_DIR/errors/"
        return
    fi

    log_info "Printing: $filename (${size} bytes)"

    # Print the file
    if lp -d "$PRINTER" "$file" 2>&1; then
        log_info "Print job submitted: $filename"

        # Move to processed folder (optional, or delete)
        mkdir -p "$WATCH_DIR/processed"
        mv "$file" "$WATCH_DIR/processed/"

        # Or delete immediately:
        # rm "$file"
    else
        log_error "Failed to print: $filename"
        mkdir -p "$WATCH_DIR/errors"
        mv "$file" "$WATCH_DIR/errors/"
    fi
}

# Watch for new files using inotifywait
log_info "Monitoring for new PDF files..."

# Process any existing files first
for file in "$WATCH_DIR"/*.pdf; do
    [ -f "$file" ] || continue
    log_info "Processing existing file: $(basename "$file")"
    print_file "$file"
done

# Watch for new files
inotifywait -m -e close_write --format '%w%f' "$WATCH_DIR" | while read -r filepath; do
    # Only process PDF files
    if [[ "$filepath" == *.pdf ]]; then
        # Wait a moment to ensure file is fully written
        sleep 1
        print_file "$filepath"
    fi
done