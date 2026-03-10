#!/bin/bash
set -e

CLUSTER_NAME="${CLUSTER_NAME:-MyCluster}"
SHARD="${SHARD:-Master}"
DST_BIN_DIR="/opt/dst_server/bin64"

echo "=== DST DS Panel Entrypoint (Linux) ==="
echo "Cluster: ${CLUSTER_NAME} | Shard: ${SHARD}"

# Always validate/update DST on start (fast if already up to date)
echo "Checking for DST server updates..."
/opt/steamcmd/steamcmd.sh +login anonymous \
    +force_install_dir /opt/dst_server \
    +app_update 343050 validate \
    +quit
echo "DST server is up to date."

DST_BIN="dontstarve_dedicated_server_nullrenderer"
if [ ! -f "${DST_BIN_DIR}/${DST_BIN}" ]; then
    DST_BIN="dontstarve_dedicated_server_nullrenderer_x64"
fi

if [ ! -f "${DST_BIN_DIR}/${DST_BIN}" ]; then
    echo "ERROR: DST server binary not found in ${DST_BIN_DIR}/"
    exit 1
fi

# Create Steam library dirs for workshop mods
mkdir -p /root/Steam/steamapps/workshop
mkdir -p /opt/dst_server/steamapps/workshop

# Copy mods setup if provided
if [ -f "/root/.klei/DoNotStarveTogether/${CLUSTER_NAME}/mods_setup.lua" ]; then
    cp "/root/.klei/DoNotStarveTogether/${CLUSTER_NAME}/mods_setup.lua" \
       /opt/dst_server/mods/dedicated_server_mods_setup.lua
fi

cd "${DST_BIN_DIR}"
echo "Starting: ./${DST_BIN} -cluster ${CLUSTER_NAME} -shard ${SHARD}"
echo "==================================="
exec "./${DST_BIN}" \
    -cluster "$CLUSTER_NAME" \
    -shard "$SHARD" \
    -console
