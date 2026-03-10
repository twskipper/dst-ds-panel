# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

DST DS Panel - A full-stack web application for managing Don't Starve Together (DST) dedicated servers via Docker containers. Features include cluster lifecycle management, real-time monitoring, server console, world settings editor, mod management, i18n, and more.

## Tech Stack

- **Backend**: Go 1.21+ (Chi router, Docker SDK, gorilla/websocket, go-ini, golang-jwt)
- **Frontend**: React 18 + TypeScript + Vite + shadcn/ui + Recharts + Monaco Editor + react-i18next
- **Container**: Docker (amd64 Debian with DST runtime libs)
- **DST Install**: DepotDownloader (macOS) or SteamCMD (Linux amd64)
- **Persistence**: JSON file store (`data/store.json`) — no database
- **Auth**: JWT tokens (24h expiry), credentials in `config.json`
- **i18n**: react-i18next with `en.json` / `zh.json`, auto-detects browser language

## Build & Run Commands

```bash
# Production: single binary with embedded frontend + world-settings.json
make build              # Output: backend/dst-ds-panel
make release            # Cross-compile: dist/dst-ds-panel-{darwin-arm64,darwin-amd64,linux-amd64}

# Development (two terminals)
make dev-backend        # Go backend on :8080 (from backend/ dir)
make dev-frontend       # Vite frontend on :5173 (proxies /api + WebSocket to :8080)

# Docker images
make docker-build       # macOS: amd64 runtime-only image, DST host-mounted
make docker-build-linux # Linux: self-contained with SteamCMD, auto-updates on start

# DST server (macOS only)
make dst-install        # Download via DepotDownloader to data/dst_server/

# CLI flags
./dst-ds-panel --dump-world-settings    # Dump embedded world-settings.json to stdout
./dst-ds-panel --world-settings file    # Use custom world-settings.json
```

## Project Structure

```
backend/
├── cmd/server/main.go          # Entry point, embed directives, reconciliation, CLI flags
├── cmd/server/frontend/        # Embedded frontend dist (gitignored, copied at build)
├── cmd/server/world-settings.json  # Embedded world settings (gitignored, copied at build)
├── internal/
│   ├── api/                    # HTTP handlers + Chi router + JWT auth
│   ├── docker/                 # Docker SDK wrapper (container lifecycle, exec, stats)
│   ├── dst/                    # DST config file I/O (cluster.ini, modoverrides.lua, templates)
│   ├── config/                 # config.json loading with env var overrides
│   ├── service/                # Background services (auto-backup, discord, health check)
│   ├── store/                  # Thread-safe JSON file persistence
│   └── model/                  # Shared data structures
├── go.mod / go.sum
frontend/
├── src/
│   ├── App.tsx                 # Root: AuthProvider, BrowserRouter, Toaster, dark mode, i18n
│   ├── pages/                  # DashboardPage, ClusterDetailPage, ClusterCreatePage, LoginPage
│   ├── components/             # UI components (see below)
│   ├── hooks/                  # useAuth, useWebSocket, useTheme
│   ├── i18n/                   # index.tsx (i18next init), en.json, zh.json
│   ├── lib/api.ts              # Typed API client with auth headers + 401 redirect
│   └── types/index.ts          # TypeScript interfaces
├── public/
│   ├── icon.png                # App icon (favicon + header + login)
│   └── world-settings.json     # World settings definitions (source of truth)
docker/
├── Dockerfile.dst              # macOS: amd64 Debian slim, runtime deps only
├── Dockerfile.linux            # Linux: includes SteamCMD, auto-updates on start
├── entrypoint.sh               # macOS entrypoint (auto-detect binary, create Steam dirs)
└── entrypoint-linux.sh         # Linux entrypoint (validate + update via SteamCMD)
deploy/
├── dst-ds-panel.service        # Systemd service file
├── docker-compose.yml          # Panel containerized deployment
└── Dockerfile.panel            # Multi-stage build for panel container
scripts/
└── install-dst.sh              # macOS: download DST via DepotDownloader
```

## Backend Architecture

### Entry Point (`cmd/server/main.go`)
- Parses CLI flags (`--dump-world-settings`, `--world-settings`)
- Loads config from `config.json` (tries current dir, then parent dir for dev)
- Resolves data directory (checks `../data` markers: `dst_server/bin64`, `clusters`, `store.json`)
- Reconciles cluster status with Docker containers on startup
- Starts background services: Discord notifier, auto-backup scheduler, health checker (30s)
- Embeds frontend via `//go:embed all:frontend` and world-settings via `//go:embed world-settings.json`
- Serves API + embedded SPA on configurable port

### API Layer (`internal/api/`)
| File | Responsibility |
|------|---------------|
| `router.go` | Chi router setup, CORS, JWT auth middleware, `/world-settings.json` serving, SPA fallback |
| `auth.go` | `POST /api/login` → JWT token; middleware checks `Authorization: Bearer` header or `?token=` query param |
| `handler_cluster.go` | CRUD, import (zip with auto root detection via `findClusterRoot`), clone (copies dir, removes token) |
| `handler_container.go` | Start (validates token exists), stop, restart; sends Discord notifications |
| `handler_console.go` | `POST .../console` → `docker exec sh -c "printf ... > /proc/1/fd/0"` with `shellQuote()` |
| `handler_logs.go` | WebSocket upgrade → Docker container logs via `stdcopy.StdCopy` demux; also handles stats streaming |
| `handler_mod.go` | Read/write `modoverrides.lua` + `mods_setup.lua` for both Master and Caves shards |
| `handler_files.go` | Read/write config files; whitelist + allowed dirs/extensions; `ListFiles` scans for extra files |
| `handler_backup.go` | Stream cluster directory as zip with `filepath.Walk` |
| `handler_players.go` | Parse `server_log.txt` for join/leave/death, `server_chat_log.txt` for chat |
| `handler_image.go` | Docker image status, DST install status + version, `UpdateDST` via DepotDownloader with log file |

### Docker Layer (`internal/docker/`)
- `NewManager(dataDir, imageName, platform)` — configurable image name and platform
- `StartShard()` — creates container with `AttachStdin: true`, `OpenStdin: true`, `NetworkMode: "host"`, `RestartPolicy: on-failure:3`
- Volume mounts: `data/clusters/{name}/ → /root/.klei/...` + `data/dst_server/ → /opt/dst_server` (if exists)
- `ExecCommand()` — writes to `/proc/1/fd/0` via `docker exec` with proper shell quoting for special chars
- `StreamLogsLines()` — uses `io.Pipe` + `stdcopy.StdCopy` to demux Docker multiplexed log stream
- `ListRunningShards()` — queries Docker by label `managed-by=dst-ds-panel` for status reconciliation

### DST Config Layer (`internal/dst/`)
- `templates.go` — full `leveldataoverride.lua` templates with all overrides for Master (forest) and Caves
- `cluster.go` — `InitClusterDir(dir, config, enableCaves)` creates cluster structure
- `modoverrides.go` — regex parser handles real-world format (enabled at end of block, not beginning)

### Background Services (`internal/service/`)
- `healthcheck.go` — every 30s checks Docker container state vs store, auto-corrects, sends Discord on crash
- `backup.go` — configurable interval, zips all clusters to `data/backups/`, keeps last 10 per cluster
- `discord.go` — sends embed messages to webhook URL on start/stop/error

## Frontend Architecture

### i18n (`src/i18n/`)
- Uses `react-i18next` with `i18next`
- Translation files: `en.json`, `zh.json` — flat nested keys like `dashboard.title`
- Interpolation: `{{name}}`, `{{shard}}` (i18next standard)
- Auto-detects browser language, persists choice to `localStorage("dst_locale")`
- `changeLanguage()` helper in `index.tsx` for language switching

### Key Components
| Component | Notes |
|-----------|-------|
| `WorldSettings` | Loads categories/options/presets from `/world-settings.json` (served by backend); parses `leveldataoverride.lua` overrides block; applies presets by merging |
| `ServerConsole` | Announce input (auto-wraps in `c_announce()`), quick action buttons, raw Lua input, command history |
| `FileEditor` | Monaco editor; loads file list from backend (`/files/list`); supports creating new files in allowed dirs |
| `LogViewer` | WebSocket connection with `?token=` auth; 500-line buffer; auto-scroll toggle |
| `StatsChart` | Two Recharts LineCharts (CPU%, Memory MB); 60-point rolling window from WebSocket |
| `ContainerControls` | Start/stop/restart with `sonner` toast notifications |

### State Management
- No global state library — uses React hooks + context
- `useAuth` — JWT token in localStorage, auto-redirect on 401
- `useWebSocket` — auto-reconnect with 3s delay, listener pattern, includes token in URL
- `useTheme` — dark class toggle on `<html>`, persists to localStorage

## Config Reference (`config.json` / `config.example.json`)

| Field | Default | Env Var | Description |
|-------|---------|---------|-------------|
| `port` | `"8080"` | `PORT` | HTTP server port |
| `dataDir` | `"./data"` | `DATA_DIR` | Data directory for clusters, backups, store |
| `imageName` | `"dst-server:latest"` | `DST_IMAGE` | Docker image for DST containers |
| `platform` | `"linux/amd64"` | `DST_PLATFORM` | Docker platform |
| `auth.username` | `"admin"` | `AUTH_USERNAME` | Login username |
| `auth.password` | `"admin"` | `AUTH_PASSWORD` | Login password |
| `auth.secret` | — | `AUTH_SECRET` | JWT signing secret (must change!) |
| `backupInterval` | `0` | — | Auto-backup hours (0=disabled) |
| `discordWebhook` | `""` | — | Discord webhook URL |

## API Routes (all require JWT except `/api/login` and `/world-settings.json`)

```
POST   /api/login                                    # Public
GET    /world-settings.json                          # Public (served by backend)
GET    /api/clusters
POST   /api/clusters                                 # Body: {name, config, enableCaves?}
POST   /api/clusters/import                          # Multipart: file + name
GET    /api/clusters/{id}
PUT    /api/clusters/{id}/config
DELETE /api/clusters/{id}
POST   /api/clusters/{id}/start                      # Requires cluster_token.txt
POST   /api/clusters/{id}/stop
POST   /api/clusters/{id}/restart
POST   /api/clusters/{id}/clone                      # Body: {name}
GET    /api/clusters/{id}/backup                     # Downloads zip
GET    /api/clusters/{id}/players                    # Parsed log events
GET    /api/clusters/{id}/mods
PUT    /api/clusters/{id}/mods
GET    /api/clusters/{id}/files?path=...
PUT    /api/clusters/{id}/files?path=...             # Body: {content}
GET    /api/clusters/{id}/files/list
GET    /api/clusters/{id}/shards/{shard}/logs        # WebSocket (token via ?token=)
GET    /api/clusters/{id}/shards/{shard}/stats       # WebSocket (token via ?token=)
POST   /api/clusters/{id}/shards/{shard}/console     # Body: {command}
GET    /api/image/status                             # {imageExists, dstInstalled, dstVersion, needsManualUpdate}
POST   /api/image/build
POST   /api/dst/update                               # Runs DepotDownloader, writes install-dst.log
```

## Important Implementation Details

### Docker Container Networking
- Uses `NetworkMode: "host"` — Master and Caves shards communicate via localhost UDP
- Multiple clusters need unique ports (default: Master=10999, Caves=10998)

### Console Command Execution
- Uses `docker exec sh -c "printf '%s\n' <shellQuoted(cmd)> > /proc/1/fd/0"`
- `shellQuote()` properly escapes single quotes for shell safety
- DST server runs as PID 1 via `exec` in entrypoint

### Zip Import Auto-Detection
- `findClusterRoot()` searches for directory containing `Master/` or `cluster.ini`
- Handles zips with nested root directories (e.g., `Cluster_2/Master/`)
- Auto-detects mods from `Master/modoverrides.lua` and generates `mods_setup.lua`

### World Settings
- Definitions stored in `frontend/public/world-settings.json` (source of truth)
- Embedded in binary via `//go:embed`, served at `/world-settings.json`
- Can be overridden with `--world-settings custom.json` flag
- Can be exported with `--dump-world-settings` for editing

### Data Directory Resolution (dev mode)
When running from `backend/` dir, checks parent `../data` for markers: `dst_server/bin64`, `clusters`, `store.json`

## Apple Silicon / macOS Notes

- DST dedicated server is x86_64 only — Docker image uses `--platform=linux/amd64`
- SteamCMD x86 binary segfaults under Docker QEMU/Rosetta emulation
- Solution: DepotDownloader on macOS host downloads Linux DST files, mounted into container
- OrbStack recommended over Docker Desktop for better Rosetta-based amd64 emulation
- `make docker-build` creates runtime-only image; `make docker-build-linux` creates self-contained image with SteamCMD

## Files NOT in Repository (gitignored)

- `config.json` — user credentials (use `config.example.json` as template)
- `data/` — runtime data (clusters, backups, store.json, dst_server, install-dst.log)
- `backend/cmd/server/frontend/` — embedded frontend build artifacts
- `backend/cmd/server/world-settings.json` — embedded world settings copy
- `dist/` — release binaries
