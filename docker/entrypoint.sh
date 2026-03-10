#!/bin/bash
set -e

echo "=== DST DS Panel Entrypoint ==="
echo "Image built: $(date -r /entrypoint.sh 2>/dev/null || echo 'unknown')"
echo "Cluster: ${CLUSTER_NAME:-MyCluster}"
echo "Shard: ${SHARD:-Master}"
echo "Architecture: $(uname -m)"
echo ""

CLUSTER_NAME="${CLUSTER_NAME:-MyCluster}"
SHARD="${SHARD:-Master}"
DST_BIN_DIR="/opt/dst_server/bin64"

# Debug: list what's in the mount
echo "Contents of /opt/dst_server/:"
ls -la /opt/dst_server/ 2>&1 || echo "  (empty or not mounted)"
echo ""
echo "Contents of ${DST_BIN_DIR}/:"
ls -la "${DST_BIN_DIR}/" 2>&1 || echo "  (directory not found)"
echo ""

# Auto-detect binary name (may be _x64 suffix depending on download method)
if [ -f "${DST_BIN_DIR}/dontstarve_dedicated_server_nullrenderer" ]; then
    DST_BIN="dontstarve_dedicated_server_nullrenderer"
elif [ -f "${DST_BIN_DIR}/dontstarve_dedicated_server_nullrenderer_x64" ]; then
    DST_BIN="dontstarve_dedicated_server_nullrenderer_x64"
else
    echo "ERROR: DST server binary not found in ${DST_BIN_DIR}/"
    echo "Install DST server first:  make dst-install"
    exit 1
fi

echo "Using binary: ${DST_BIN}"
echo "Binary info: $(file ${DST_BIN_DIR}/${DST_BIN} 2>/dev/null || echo 'unknown')"
echo ""

# Create Steam library directories required for workshop mod downloads
mkdir -p /root/Steam/steamapps/workshop
mkdir -p /root/steamapps/workshop
mkdir -p /opt/dst_server/steamapps/workshop
# Symlink so DST can find the staging/install library folder
ln -sf /opt/dst_server/steamapps /root/.steam/steamapps 2>/dev/null || true
mkdir -p /root/.steam/steam/steamapps
ln -sf /opt/dst_server/steamapps/workshop /root/.steam/steam/steamapps/workshop 2>/dev/null || true

# Copy mods setup if provided
if [ -f "/root/.klei/DoNotStarveTogether/${CLUSTER_NAME}/mods_setup.lua" ]; then
    cp "/root/.klei/DoNotStarveTogether/${CLUSTER_NAME}/mods_setup.lua" \
       /opt/dst_server/mods/dedicated_server_mods_setup.lua
fi

cd "${DST_BIN_DIR}"
echo "Starting DST server: ./${DST_BIN} -cluster ${CLUSTER_NAME} -shard ${SHARD}"
echo "==================================="
exec "./${DST_BIN}" \
    -cluster "$CLUSTER_NAME" \
    -shard "$SHARD" \
    -console
