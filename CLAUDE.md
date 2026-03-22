# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

DST DS Panel - A cross-platform web application for managing Don't Starve Together (DST) dedicated servers. On Windows, DST runs as native processes (no Docker). On macOS/Linux, DST runs in Docker containers. Users can create/import server clusters, configure mods, start/stop servers, monitor logs/resources, send console commands, and manage network ports.

## Tech Stack

- **Backend**: Go (Chi router, Docker SDK, gorilla/websocket, go-ini, golang-jwt)
- **Frontend**: React + TypeScript + Vite + shadcn/ui + Recharts + Monaco Editor
- **Runtime**: Native processes (Windows) or Docker containers (macOS/Linux)
- **DST Install**: DepotDownloader (macOS/Windows auto-download) or SteamCMD (Linux Docker)
- **Persistence**: JSON file store (`data/store.json`)
- **Auth**: JWT-based login, credentials in `config.json`

## Build & Run Commands

```bash
# Production: single binary with embedded frontend
make build              # Output: backend/dst-ds-panel
make release            # Cross-compile: macOS DMG, Windows zip, Linux binary
make release-windows    # Windows only: dist/DST-DS-Panel-windows-x64.zip

# Development
make dev-backend        # Go backend on :8080
make dev-frontend       # Vite frontend on :5173 (proxies /api to :8080)

# Docker images
make docker-build       # macOS: runtime-only image (DST mounted from host)
make docker-build-linux # Linux: self-contained with SteamCMD

# DST server (macOS only)
make dst-install        # Download via DepotDownloader
```

## Architecture

### Backend (`backend/`)

- **Entry**: `cmd/server/main.go` - starts server, auto-detects mode (native on Windows, docker on macOS/Linux), reconciles state on startup
- **Entry**: `cmd/tray/main.go` - system tray app (Windows/macOS), manages server process lifecycle
- **`internal/api/`** - HTTP handlers and Chi router
  - `router.go` - route definitions, JWT auth middleware, embedded frontend SPA serving, Handler with ShardManager interface
  - `auth.go` - login endpoint + JWT middleware (token in Authorization header or ?token= for WebSocket)
  - `handler_cluster.go` - cluster CRUD, import (zip with auto root detection), clone, port config (GET/PUT /ports), auto port assignment
  - `handler_container.go` - start/stop/restart with token validation and Discord notifications
  - `handler_console.go` - send Lua commands to running server via ShardManager
  - `handler_logs.go` - WebSocket log/stats streaming
  - `handler_mod.go` - mod list/update (writes modoverrides.lua + mods_setup.lua)
  - `handler_files.go` - read/write config files (whitelisted paths + allowed extensions)
  - `handler_backup.go` - stream cluster as zip download
  - `handler_players.go` - parse server logs for join/leave/chat/death events
  - `handler_image.go` - DST install status, update via DepotDownloader (auto-download on Windows), branch selection
- **`internal/manager/`** - ShardManager interface abstraction
  - `manager.go` - `ShardManager` interface (StartShard, StopShard, ExecCommand, StreamLogs, etc.)
  - `docker.go` - Docker adapter wrapping `docker.Manager`
  - `process.go` - Native process manager (Windows) — runs DST as OS processes with stdin/stdout pipes
- **`internal/docker/`** - Docker SDK wrapper (used by DockerManager)
  - Container lifecycle with `AttachStdin`/`OpenStdin` for console commands
  - `ExecCommand()` writes to `/proc/1/fd/0` with proper shell quoting
  - `ListRunningShards()` for status reconciliation on startup
- **`internal/dst/`** - DST config file I/O
  - `cluster.go` - cluster.ini read/write, InitClusterDir with optional caves, ReadMasterPort/WriteMasterPort, ReadShardPort/WriteShardPort
  - `modoverrides.go` - parse/generate modoverrides.lua (handles real-world multi-line format)
  - `templates.go` - full leveldataoverride.lua templates for Master and Caves
- **`internal/config/`** - config.json with auth, backup interval, discord webhook
- **`internal/service/`** - auto-backup scheduler, Discord webhook notifications
- **`internal/store/`** - thread-safe JSON file persistence
- **`internal/model/`** - shared data structures

### Frontend (`frontend/src/`)

- **Pages**:
  - `DashboardPage` - cluster cards, DST status badge, Update DST button (macOS only), auto-refresh 10s
  - `ClusterCreatePage` - new cluster form with caves toggle, import zip tab
  - `ClusterDetailPage` - tabs: Overview, Master, Caves, Console, World, Mods, Files
  - `LoginPage` - JWT login with icon
- **Components**:
  - `ClusterCard` - status badge, quick start/stop with error alerts
  - `ClusterForm` - cluster.ini editor with token help text
  - `ContainerControls` - start/stop/restart with error alerts
  - `LogViewer` - WebSocket real-time logs with auto-scroll
  - `StatsChart` - Recharts CPU/memory line charts via WebSocket
  - `ServerConsole` - announce input, quick actions (save/rollback/regenerate/shutdown), raw Lua command
  - `WorldSettings` - visual editor for leveldataoverride.lua with difficulty presets (Easy/Normal/Hard/Challenge) and event settings
  - `FileEditor` - Monaco editor with Lua/INI highlighting, file creation dialog
  - `ModList` + `ModConfigDialog` - mod management by Workshop ID
  - `AdminList` - manage adminlist.txt
  - `PlayerActivity` - parsed join/leave/chat/death events from server logs
- **Hooks**: `useAuth` (JWT context), `useWebSocket` (auto-reconnect with token), `useTheme` (dark mode)
- **API client**: `lib/api.ts` - typed fetch with auto 401→login redirect

### Docker (`docker/`)

- `Dockerfile.dst` - macOS: amd64 Debian slim, runtime deps only, DST host-mounted
- `Dockerfile.linux` - Linux: includes SteamCMD, auto-updates DST on each start
- `entrypoint.sh` / `entrypoint-linux.sh` - auto-detect binary name, create Steam dirs for mod downloads

### Config (`config.json`)

```json
{
  "port": "8080",
  "dataDir": "./data",
  "imageName": "dst-server:latest",
  "platform": "linux/amd64",
  "auth": { "username": "admin", "password": "admin", "secret": "change-me" },
  "backupInterval": 6,
  "discordWebhook": "https://discord.com/api/webhooks/..."
}
```

All fields overridable via env vars: `PORT`, `DATA_DIR`, `DST_IMAGE`, `DST_PLATFORM`, `AUTH_USERNAME`, `AUTH_PASSWORD`, `AUTH_SECRET`.

## API Routes

```
POST   /api/login                              # Public: JWT login
GET    /api/clusters                           # List clusters
POST   /api/clusters                           # Create (with enableCaves option)
POST   /api/clusters/import                    # Import zip (auto root detection + mod detection)
GET    /api/clusters/{id}                      # Get cluster
PUT    /api/clusters/{id}/config               # Update config
DELETE /api/clusters/{id}                      # Delete
POST   /api/clusters/{id}/start               # Start (requires token)
POST   /api/clusters/{id}/stop                # Stop
POST   /api/clusters/{id}/restart             # Restart
POST   /api/clusters/{id}/clone               # Clone cluster
GET    /api/clusters/{id}/backup              # Download zip
GET    /api/clusters/{id}/players             # Player activity from logs
GET    /api/clusters/{id}/mods                # List mods
PUT    /api/clusters/{id}/mods                # Update mods
GET    /api/clusters/{id}/files?path=...      # Read config file
PUT    /api/clusters/{id}/files?path=...      # Write config file
GET    /api/clusters/{id}/files/list           # List editable files
GET    /api/clusters/{id}/shards/{shard}/logs  # WebSocket logs
GET    /api/clusters/{id}/shards/{shard}/stats # WebSocket CPU/mem
POST   /api/clusters/{id}/shards/{shard}/console # Send Lua command
GET    /api/image/status                       # DST install status
POST   /api/image/build                        # Build Docker image
GET    /api/clusters/{id}/ports                # Get shard ports + master_port
PUT    /api/clusters/{id}/ports                # Update ports (with conflict check)
POST   /api/dst/update                         # Update DST (?branch= for beta)
```

## Apple Silicon Notes

- DST dedicated server is x86_64 only - Docker image uses `--platform=linux/amd64`
- SteamCMD segfaults under Docker emulation - use DepotDownloader on host
- OrbStack recommended over Docker Desktop for Rosetta-based amd64 emulation
