#!/bin/bash
set -e

# Accept custom path as argument, or use relative to script location
if [ -n "$1" ]; then
    DST_DIR="$1"
else
    DST_DIR="$(cd "$(dirname "$0")/.." && pwd)/data/dst_server"
fi
APP_ID=343050

echo "=== DST Dedicated Server Installer ==="
echo "Install directory: ${DST_DIR}"
echo ""

# Check if already installed
if [ -f "${DST_DIR}/bin64/dontstarve_dedicated_server_nullrenderer" ] || \
   [ -f "${DST_DIR}/bin64/dontstarve_dedicated_server_nullrenderer_x64" ]; then
    read -p "DST server already installed. Re-install/update? (y/N): " REPLY
    if [ "$REPLY" != "y" ] && [ "$REPLY" != "Y" ]; then
        echo "Skipping."
        exit 0
    fi
fi

mkdir -p "${DST_DIR}"

# Check for DepotDownloader
if ! command -v DepotDownloader &> /dev/null; then
    echo "DepotDownloader not found. Installing via Homebrew..."
    if ! command -v brew &> /dev/null; then
        echo "ERROR: Homebrew not found. Install from https://brew.sh"
        exit 1
    fi
    brew install SteamRE/tools/depotdownloader
fi

echo ""
echo "Downloading DST Dedicated Server (Linux) via DepotDownloader..."
echo "App ID: ${APP_ID}"
echo "This will download ~2GB of files..."
echo ""

DepotDownloader -app ${APP_ID} -os linux -dir "${DST_DIR}"

# Fix permissions (DepotDownloader doesn't preserve execute bits)
chmod +x "${DST_DIR}/bin64/"dontstarve_dedicated_server_nullrenderer* 2>/dev/null || true
find "${DST_DIR}/bin64" -name "*.so*" -exec chmod +x {} \; 2>/dev/null || true

echo ""
echo "=== DST server installed successfully! ==="
echo "Location: ${DST_DIR}"
