<p align="center">
  <img src="frontend/public/icon.png" alt="DST DS Panel" width="128" height="128">
</p>

<h1 align="center">DST DS Panel</h1>

<p align="center">
  <a href="README.md">English</a> | <a href="README_zh.md">中文</a>
</p>

<p align="center">
  A web-based management panel for Don't Starve Together dedicated servers.<br>
  Create, configure, and monitor multiple DST server clusters through a modern UI.
</p>

## Features

- **Cluster Management** — Create new worlds (with optional Caves), import existing saves, clone clusters
- **Server Console** — Send Lua commands (save, rollback, announce, regenerate) directly from the UI
- **Config File Editor** — Monaco editor with Lua/INI syntax highlighting
- **Mod Management** — Add mods by Workshop ID, auto-detect mods from imported saves
- **Admin Management** — Manage server admins (adminlist.txt)
- **Player Activity** — View player join/leave, chat, and death events parsed from server logs
- **Port Management** — Configure network ports per cluster with auto conflict detection
- **Live Logs** — Real-time log streaming from Master and Caves shards
- **Resource Monitoring** — Live CPU and memory charts per shard
- **Backup & Restore** — Download cluster backups as zip, import to restore
- **Auto-Backup** — Scheduled backups every N hours (configurable)
- **Discord Notifications** — Webhook alerts for server start/stop/error
- **Multi-Cluster** — Run multiple server worlds simultaneously with auto port assignment
- **Beta Branch** — Install DST stable or beta versions from the dashboard
- **Dark Mode** — Toggle between light and dark themes
- **Authentication** — Login system with JWT tokens
- **Single Binary** — Frontend embedded in Go binary, no separate web server needed
- **Cross-Platform** — Windows (native), macOS (Docker), Linux (Docker)

---

## Setup — Windows (Recommended)

The simplest way to run DST dedicated servers. No Docker required.

### Step 1: Download

Download `DST-DS-Panel-windows-x64.zip` from the [Releases](../../releases) page and extract to a folder (e.g. `C:\DST-DS-Panel`).

### Step 2: Run

Double-click `dst-ds-panel-tray.exe`. A **DST** icon appears in the system tray (bottom-right).

Click the tray icon → **Start Server** → **Open Panel**.

### Step 3: Install DST

In the web panel, click **"Install DST"** on the dashboard. DepotDownloader will be auto-downloaded and DST dedicated server will be installed (~2GB).

### Step 4: Create & Play

1. Click **"New Cluster"**
2. Fill in server name, paste your [cluster token](#1-get-a-cluster-token)
3. Click **"Create Cluster"** → **"Start Server"**

Data is stored next to the exe. Fully portable — move the folder anywhere.

---

## Setup — macOS (Apple Silicon)

### Prerequisites

- [Docker Desktop](https://www.docker.com/products/docker-desktop/) or [OrbStack](https://orbstack.dev/) (recommended)
- [Homebrew](https://brew.sh/)

### Option A: macOS App (Recommended)

Download `DST.DS.Panel.dmg` from the [Releases](../../releases) page, open the DMG, and drag the app to `/Applications`.

If macOS blocks the app, use one of these fixes:
- Double-click `Fix Permissions.command` in the DMG
- Or run in Terminal: `xattr -cr "/Applications/DST DS Panel.app"`
- Or go to **System Settings → Privacy & Security**, find "DST DS Panel was blocked", click **Open Anyway**

Double-click to launch — it runs as a menu bar app with one-click server start/stop.

### Option B: Binary

Download `dst-ds-panel-darwin-arm64` from [Releases](../../releases).

```bash
chmod +x dst-ds-panel-darwin-arm64
./dst-ds-panel-darwin-arm64
```

Open `http://localhost:8080` and login (default: admin/admin).

Click **"Install DST"** on the dashboard to download the DST server. The Docker image will be auto-pulled when you start your first cluster.

---

## Quick Start — Docker (Linux)

The fastest way to get started on any Linux server:

```bash
docker run -d \
  --name dst-ds-panel \
  -p 8080:8080 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v dst-panel-data:/app/data \
  -e AUTH_PASSWORD=your-password \
  -e AUTH_SECRET=your-random-secret \
  -e DST_IMAGE=twskipper/dst-ds-runtime:linux \
  --restart unless-stopped \
  twskipper/dst-ds-panel:latest
```

Or use Docker Compose:

```bash
curl -O https://raw.githubusercontent.com/twskipper/dst-ds-panel/main/deploy/docker-compose.yml
curl -O https://raw.githubusercontent.com/twskipper/dst-ds-panel/main/config.example.json
cp config.example.json config.json
# Edit config.json to set your password and secret
docker compose up -d
```

Open `http://your-server:8080` and login (default: admin/admin).

> **Note:** Docker socket mount (`/var/run/docker.sock`) is required for the panel to manage DST containers. The DST runtime image will be pulled automatically when you start your first cluster.

---

## Setup — Linux amd64 (Binary)

### Prerequisites

- [Docker](https://docs.docker.com/engine/install/)

Download `dst-ds-panel-linux-amd64` from [Releases](../../releases).

```bash
chmod +x dst-ds-panel-linux-amd64
./dst-ds-panel-linux-amd64
```

Open `http://your-server:8080` and login. The Docker image will be auto-pulled on first cluster start.

---

## Using the Panel

### 1. Get a Cluster Token

Before starting any server, you need a Klei cluster token:

1. Go to [Klei Account — Game Servers](https://accounts.klei.com/account/game/servers?game=DontStarveTogether)
2. Click **"Add New Server"**
3. Copy the generated token

Or in-game, press `~` and run `TheNet:GenerateClusterToken()`.

### 2. Create a New Cluster

1. Click **"New Cluster"** on the dashboard
2. Fill in server name, game mode, max players
3. Toggle **"Enable Caves"** on/off
4. Paste your cluster token
5. Click **"Create Cluster"**

Network ports are auto-assigned to avoid conflicts when running multiple clusters.

### 3. Import an Existing Save

1. Click **"New Cluster"** → **"Import"** tab
2. Upload a zip file of your cluster directory (handles nested folders automatically)
3. Mods are automatically detected from `modoverrides.lua`

### 4. Start the Server

1. Click **"Start Server"** on the cluster detail page (requires cluster token)
2. First start may take a few minutes (downloading mods)
3. View live logs in the **Master** and **Caves** tabs

### 5. Manage the Server

- **Overview** — Edit game settings, configure network ports, manage admins, view player activity
- **Console** — Send commands (save, rollback, announce, regenerate), raw Lua input
- **Mods** — Add/remove mods by Workshop ID, edit per-mod configuration
- **Files** — Edit any config file with Monaco editor (Lua/INI highlighting)
- **Backup** — Download cluster as zip
- **Clone** — Duplicate cluster with a new name

### 6. Update DST Server

Select **Stable**, **Beta**, or **Previous Update** from the dropdown on the dashboard, then click **"Update DST"**.

---

## Configuration

### config.json

| Field | Default | Description |
|-------|---------|-------------|
| `port` | `"8080"` | HTTP server port |
| `dataDir` | `"./data"` | Directory for clusters, saves, and state |
| `imageName` | `"dst-server:latest"` | Docker image name (Docker mode only) |
| `platform` | `"linux/amd64"` | Docker platform (Docker mode only) |
| `auth.username` | `"admin"` | Login username |
| `auth.password` | `"admin"` | Login password |
| `auth.secret` | — | JWT signing secret (change this!) |
| `backupInterval` | `0` | Auto-backup interval in hours (0 = disabled) |
| `discordWebhook` | `""` | Discord webhook URL for server notifications |

All fields can be overridden with environment variables: `PORT`, `DATA_DIR`, `DST_IMAGE`, `DST_PLATFORM`, `AUTH_USERNAME`, `AUTH_PASSWORD`, `AUTH_SECRET`.

Runtime mode is auto-detected: **native** on Windows (no Docker), **docker** on macOS/Linux. Override with `DST_MODE=native` or `DST_MODE=docker`.

---

## Development

Building from source requires [Go](https://go.dev/) 1.21+ and [Node.js](https://nodejs.org/) 18+.

```bash
# Install frontend dependencies
cd frontend && npm install && cd ..

# Development (two terminals)
make dev-backend     # Go backend on :8080
make dev-frontend    # Vite frontend on :5173 (proxies /api to :8080)

# Production build (single binary with embedded frontend)
make build           # Output: backend/dst-ds-panel

# Cross-compile for all platforms
make release         # Output: dist/ (macOS DMG, Windows zip, Linux binary)
make release-windows # Windows only
```

## Architecture

```
                          ┌─── Docker Mode (macOS/Linux) ───┐
Browser ──HTTP/WS──→ Go  │  Docker SDK → containers        │
                   Backend│                                  │
                          ├─── Native Mode (Windows) ───────┤
                          │  OS processes (no Docker)        │
                          └──────────────────────────────────┘

Data: clusters/{name}/    DST server: dst_server/bin64/
```

- **Windows**: DST runs as native processes, no Docker required
- **macOS/Linux**: Each shard runs as a Docker container with host networking
- Config files and saves stored in `data/clusters/`
- Ports auto-assigned per cluster to avoid conflicts in multi-server setups
