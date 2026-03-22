<p align="center">
  <img src="frontend/public/icon.png" alt="DST DS Panel" width="128" height="128">
</p>

<h1 align="center">DST 服务器面板</h1>

<p align="center">
  <a href="README.md">English</a> | <a href="README_zh.md">中文</a>
</p>

<p align="center">
  基于 Web 的饥荒联机版（Don't Starve Together）专用服务器管理面板。<br>
  通过现代化 UI 创建、配置和监控多个 DST 服务器集群。
</p>

## 功能特性

- **集群管理** — 创建新世界（可选洞穴）、导入现有存档、克隆集群
- **服务器控制台** — 直接从 UI 发送 Lua 命令（保存、回档、公告、重新生成世界）
- **配置文件编辑器** — Monaco 编辑器，支持 Lua/INI 语法高亮
- **模组管理** — 通过创意工坊 ID 添加模组，导入时自动检测已有模组
- **管理员管理** — 管理服务器管理员列表（adminlist.txt）
- **玩家活动** — 查看玩家加入/离开、聊天和死亡事件
- **端口管理** — 每个集群独立配置网络端口，自动冲突检测
- **实时日志** — Master 和 Caves 分片的实时日志流
- **资源监控** — 每个分片的实时 CPU 和内存图表
- **备份与恢复** — 下载集群备份 zip，导入恢复
- **自动备份** — 可配置每 N 小时自动备份
- **Discord 通知** — 服务器启动/停止/错误时发送 Webhook 通知
- **多集群** — 同时运行多个服务器世界，端口自动分配
- **测试版分支** — 从面板安装 DST 正式版或测试版
- **深色模式** — 浅色/深色主题切换
- **多语言** — 中文/英文界面切换
- **登录认证** — 基于 JWT 的登录系统
- **单一二进制** — 前端嵌入 Go 二进制，无需独立 Web 服务器
- **跨平台** — Windows（原生）、macOS（Docker）、Linux（Docker）

---

## 安装 — Windows（推荐）

最简单的 DST 专服搭建方式，无需 Docker。

### 步骤 1：下载

从 [Releases](../../releases) 页面下载 `DST-DS-Panel-windows-x64.zip`，解压到任意文件夹（如 `C:\DST-DS-Panel`）。

### 步骤 2：运行

双击 `dst-ds-panel-tray.exe`，系统托盘（右下角）会出现 **DST** 图标。

点击托盘图标 → **Start Server** → **Open Panel**。

### 步骤 3：安装 DST

在面板中点击仪表盘上的 **"安装 DST"**，DepotDownloader 会自动下载，然后安装 DST 专用服务器（约 2GB）。

### 步骤 4：创建并开玩

1. 点击 **"新建集群"**
2. 填写服务器名称，粘贴[集群令牌](#1-获取集群令牌)
3. 点击 **"创建集群"** → **"启动服务器"**

数据存储在 exe 同目录下，完全便携 — 随意移动文件夹即可。

---

## 安装 — macOS (Apple Silicon)

### 前置要求

- [Docker Desktop](https://www.docker.com/products/docker-desktop/) 或 [OrbStack](https://orbstack.dev/)（推荐）
- [Homebrew](https://brew.sh/)

### 方式 A：macOS App（推荐）

从 [Releases](../../releases) 页面下载 `DST.DS.Panel.dmg`，打开后拖到 `/Applications`。

如果 macOS 阻止打开，使用以下任一方法：
- 双击 DMG 中的 `Fix Permissions.command` 脚本
- 或在终端运行：`xattr -cr "/Applications/DST DS Panel.app"`
- 或前往 **系统设置 → 隐私与安全性**，找到"DST DS Panel 已被阻止"，点击 **仍要打开**

双击启动，作为菜单栏应用运行，一键启停服务器。

### 方式 B：二进制文件

从 [Releases](../../releases) 下载 `dst-ds-panel-darwin-arm64`。

```bash
chmod +x dst-ds-panel-darwin-arm64
./dst-ds-panel-darwin-arm64
```

打开 `http://localhost:8080` 登录（默认：admin/admin）。

点击仪表盘上的 **"安装 DST"** 下载 DST 服务器。首次启动集群时 Docker 镜像会自动拉取。

---

## 快速启动 — Docker 一键部署（Linux）

在任何 Linux 服务器上最快的启动方式：

```bash
docker run -d \
  --name dst-ds-panel \
  -p 8080:8080 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v dst-panel-data:/app/data \
  -e AUTH_PASSWORD=你的密码 \
  -e AUTH_SECRET=随机密钥字符串 \
  -e DST_IMAGE=twskipper/dst-ds-runtime:linux \
  --restart unless-stopped \
  twskipper/dst-ds-panel:latest
```

或使用 Docker Compose：

```bash
curl -O https://raw.githubusercontent.com/twskipper/dst-ds-panel/main/deploy/docker-compose.yml
curl -O https://raw.githubusercontent.com/twskipper/dst-ds-panel/main/config.example.json
cp config.example.json config.json
# 编辑 config.json 设置密码
docker compose up -d
```

打开 `http://你的服务器:8080` 登录（默认：admin/admin）。

> **注意：** 需要挂载 Docker socket（`/var/run/docker.sock`）。DST 运行时镜像会在首次启动集群时自动拉取。

---

## 安装 — Linux amd64（二进制）

### 前置要求

- [Docker](https://docs.docker.com/engine/install/)

从 [Releases](../../releases) 下载 `dst-ds-panel-linux-amd64`。

```bash
chmod +x dst-ds-panel-linux-amd64
./dst-ds-panel-linux-amd64
```

打开 `http://你的服务器:8080` 登录。首次启动集群时 Docker 镜像会自动拉取。

---

## 使用指南

### 1. 获取集群令牌

启动服务器前需要 Klei 集群令牌：

1. 访问 [Klei 账户 — 游戏服务器](https://accounts.klei.com/account/game/servers?game=DontStarveTogether)
2. 点击 **"Add New Server"**
3. 复制生成的令牌

或在游戏中按 `~` 键运行 `TheNet:GenerateClusterToken()`。

### 2. 创建新集群

1. 点击仪表盘上的 **"新建集群"**
2. 填写服务器名称、游戏模式、最大玩家数
3. 选择是否 **启用洞穴**
4. 粘贴集群令牌
5. 点击 **"创建集群"**

运行多个集群时，网络端口会自动分配以避免冲突。

### 3. 导入现有存档

1. 点击 **"新建集群"** → **"导入"** 标签
2. 上传集群目录的 zip 文件（自动处理嵌套文件夹）
3. 模组从 `modoverrides.lua` 自动检测

### 4. 启动服务器

1. 在集群详情页点击 **"启动服务器"**（需要集群令牌）
2. 首次启动可能需要几分钟（下载模组）
3. 在 **Master** 和 **Caves** 标签中查看实时日志

### 5. 管理服务器

- **概览** — 编辑游戏设置、配置网络端口、管理管理员、查看玩家活动
- **控制台** — 发送命令（保存、回档、公告、重新生成世界）、原始 Lua 输入
- **模组** — 通过创意工坊 ID 添加/移除模组，编辑模组配置
- **文件** — 使用 Monaco 编辑器编辑任何配置文件（Lua/INI 高亮）
- **备份** — 下载集群 zip 备份
- **克隆** — 复制集群并使用新名称

### 6. 更新 DST 服务器

在仪表盘的下拉菜单中选择 **正式版**、**测试版** 或 **上一版本**，然后点击 **"更新 DST"**。

---

## 配置

### config.json

| 字段 | 默认值 | 说明 |
|------|--------|------|
| `port` | `"8080"` | HTTP 服务端口 |
| `dataDir` | `"./data"` | 集群、存档和状态数据目录 |
| `imageName` | `"dst-server:latest"` | DST 容器的 Docker 镜像名（仅 Docker 模式） |
| `platform` | `"linux/amd64"` | Docker 平台（仅 Docker 模式） |
| `auth.username` | `"admin"` | 登录用户名 |
| `auth.password` | `"admin"` | 登录密码 |
| `auth.secret` | — | JWT 签名密钥（务必修改！） |
| `backupInterval` | `0` | 自动备份间隔（小时），0 为禁用 |
| `discordWebhook` | `""` | Discord Webhook URL |

所有字段可通过环境变量覆盖：`PORT`, `DATA_DIR`, `DST_IMAGE`, `DST_PLATFORM`, `AUTH_USERNAME`, `AUTH_PASSWORD`, `AUTH_SECRET`。

运行模式自动检测：Windows 使用 **原生模式**（无需 Docker），macOS/Linux 使用 **Docker 模式**。可通过 `DST_MODE=native` 或 `DST_MODE=docker` 覆盖。

---

## 开发

从源码构建需要 [Go](https://go.dev/) 1.21+ 和 [Node.js](https://nodejs.org/) 18+。

```bash
# 安装前端依赖
cd frontend && npm install && cd ..

# 开发模式（两个终端）
make dev-backend     # Go 后端 :8080
make dev-frontend    # Vite 前端 :5173（代理 /api 到 :8080）

# 生产构建（前端嵌入的单一二进制）
make build           # 输出：backend/dst-ds-panel

# 跨平台编译
make release         # 输出：dist/（macOS DMG、Windows zip、Linux 二进制）
make release-windows # 仅 Windows
```

## 架构

```
                          ┌─── Docker 模式 (macOS/Linux) ───┐
浏览器 ──HTTP/WS──→ Go   │  Docker SDK → 容器              │
                   后端   │                                  │
                          ├─── 原生模式 (Windows) ──────────┤
                          │  操作系统进程（无需 Docker）      │
                          └──────────────────────────────────┘

数据：clusters/{名称}/    DST 服务器：dst_server/bin64/
```

- **Windows**：DST 以原生进程运行，无需 Docker
- **macOS/Linux**：每个分片作为 Docker 容器运行，使用 host 网络
- 配置文件和存档存储在 `data/clusters/`
- 多服务器运行时端口自动分配，避免冲突
