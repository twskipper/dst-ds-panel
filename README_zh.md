<p align="center">
  <img src="frontend/public/icon.png" alt="DST DS Panel" width="128" height="128">
</p>

<h1 align="center">DST 服务器面板</h1>

<p align="center">
  <a href="README.md">English</a> | <a href="README_zh.md">中文</a>
</p>

<p align="center">
  基于 Web 的饥荒联机版（Don't Starve Together）专用服务器管理面板。<br>
  通过现代化 UI 创建、配置和监控多个 DST 服务器集群，<br>
  每个集群以 Docker 容器方式运行。
</p>

## 功能特性

- **集群管理** — 创建新世界（可选洞穴）、导入现有存档、克隆集群
- **世界设置 UI** — 可视化世界生成编辑器，内置难度预设（简单/适中/困难/挑战）和节庆活动开关
- **服务器控制台** — 直接从 UI 发送 Lua 命令（保存、回档、公告、重新生成世界）
- **配置文件编辑器** — Monaco 编辑器，支持 Lua/INI 语法高亮
- **模组管理** — 通过创意工坊 ID 添加模组，导入时自动检测已有模组
- **管理员管理** — 管理服务器管理员列表（adminlist.txt）
- **玩家活动** — 查看玩家加入/离开、聊天和死亡事件
- **容器生命周期** — 一键启动、停止、重启服务器
- **实时日志** — Master 和 Caves 分片的实时日志流
- **资源监控** — 每个分片的实时 CPU 和内存图表
- **备份与恢复** — 下载集群备份 zip，导入恢复
- **自动备份** — 可配置每 N 小时自动备份
- **Discord 通知** — 服务器启动/停止/错误时发送 Webhook 通知
- **多集群** — 同时运行多个服务器世界
- **深色模式** — 浅色/深色主题切换
- **多语言** — 中文/英文界面切换
- **登录认证** — 基于 JWT 的登录系统
- **自动更新** — 从面板更新 DST 服务器（macOS）或容器启动时自动更新（Linux）
- **单一二进制** — 前端嵌入 Go 二进制，无需独立 Web 服务器

---

## 安装 — macOS (Apple Silicon)

### 前置要求

- [Docker Desktop](https://www.docker.com/products/docker-desktop/) 或 [OrbStack](https://orbstack.dev/)（推荐）
- [Homebrew](https://brew.sh/) — 用于安装 DepotDownloader

### 步骤 1：下载

**方式 A：macOS App（推荐）**

从 [Releases](../../releases) 页面下载 `DST.DS.Panel.dmg`，打开后拖到 `/Applications`。

如果 macOS 阻止打开，使用以下任一方法：
- 双击 DMG 中的 `Fix Permissions.command` 脚本
- 或在终端运行：`xattr -cr "/Applications/DST DS Panel.app"`
- 或前往 **系统设置 → 隐私与安全性**，找到"DST DS Panel 已被阻止"，点击 **仍要打开**

双击启动，作为菜单栏应用运行，一键启停服务器。

**方式 B：二进制文件**

从 [Releases](../../releases) 页面下载 `dst-ds-panel-darwin-arm64`。

```bash
chmod +x dst-ds-panel-darwin-arm64
```

### 步骤 2：构建 Docker 镜像

```bash
docker build --platform linux/amd64 -f docker/Dockerfile.dst -t dst-server:latest docker/
```

### 步骤 3：安装 DST 服务器

```bash
./scripts/install-dst.sh
```

或者启动面板后点击仪表盘上的 **"更新 DST"** 按钮。

### 步骤 4：配置

创建 `config.json`：

```json
{
  "port": "8080",
  "dataDir": "./data",
  "imageName": "dst-server:latest",
  "platform": "linux/amd64",
  "auth": {
    "username": "admin",
    "password": "修改为你的密码",
    "secret": "修改为随机字符串"
  },
  "backupInterval": 6,
  "discordWebhook": ""
}
```

### 步骤 5：运行

```bash
./dst-ds-panel-darwin-arm64
```

打开 `http://localhost:8080` 登录。

---

## 快速启动 — Docker 一键部署

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

打开 `http://你的服务器:8080` 登录（默认：admin/change-me）。

> **注意：** 需要挂载 Docker socket（`/var/run/docker.sock`）以便面板管理 DST 容器。DST 运行时镜像（`twskipper/dst-ds-runtime:linux`）会在首次启动集群时自动拉取。

---

## 安装 — Linux amd64

### 前置要求

- [Docker](https://docs.docker.com/engine/install/)

### 步骤 1：下载

从 [Releases](../../releases) 页面下载 `dst-ds-panel-linux-amd64`。

```bash
chmod +x dst-ds-panel-linux-amd64
```

### 步骤 2：构建 Docker 镜像

```bash
docker build -f docker/Dockerfile.linux -t dst-server:latest docker/
```

此镜像包含 SteamCMD，容器首次启动时会自动下载和更新 DST 服务器。

### 步骤 3：配置

创建 `config.json`（同上）。

### 步骤 4：运行

```bash
./dst-ds-panel-linux-amd64
```

打开 `http://你的服务器:8080` 登录。

### Systemd 服务（可选）

```bash
sudo cp deploy/dst-ds-panel.service /etc/systemd/system/
sudo systemctl enable dst-ds-panel
sudo systemctl start dst-ds-panel
```

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

### 3. 管理服务器

- **概览** — 编辑游戏设置、管理管理员、查看玩家活动
- **控制台** — 发送命令（保存、回档、公告、重新生成世界）、原始 Lua 输入
- **世界设置** — 可视化世界设置编辑器，内置难度预设和节庆活动开关
- **模组** — 通过创意工坊 ID 添加/移除模组，编辑模组配置
- **文件** — 使用 Monaco 编辑器编辑任何配置文件（Lua/INI 高亮）
- **备份** — 下载集群 zip 备份
- **克隆** — 复制集群并使用新名称

---

## 配置

### config.json

| 字段 | 默认值 | 说明 |
|------|--------|------|
| `port` | `"8080"` | HTTP 服务端口 |
| `dataDir` | `"./data"` | 集群、存档和状态数据目录 |
| `imageName` | `"dst-server:latest"` | DST 容器的 Docker 镜像名 |
| `platform` | `"linux/amd64"` | Docker 平台 |
| `auth.username` | `"admin"` | 登录用户名 |
| `auth.password` | `"admin"` | 登录密码 |
| `auth.secret` | — | JWT 签名密钥（务必修改！） |
| `backupInterval` | `0` | 自动备份间隔（小时），0 为禁用 |
| `discordWebhook` | `""` | Discord Webhook URL |

所有字段可通过环境变量覆盖：`PORT`, `DATA_DIR`, `DST_IMAGE`, `DST_PLATFORM`, `AUTH_USERNAME`, `AUTH_PASSWORD`, `AUTH_SECRET`。

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
make release         # 输出：dist/dst-ds-panel-{darwin-arm64,darwin-amd64,linux-amd64}
```

## 架构

```
浏览器 ──HTTP/WS──→ Go 后端 ──Docker SDK──→ dst-{集群}-master（容器）
                        │                  → dst-{集群}-caves（容器）
                        │
                   data/clusters/{名称}/    （卷挂载到容器中）
                   data/dst_server/         （DST 二进制，挂载或内置）
```

- 每个集群的 Master 和 Caves 分片作为独立 Docker 容器运行
- 容器使用 **host 网络模式** 实现分片间 UDP 通信
- 配置文件和存档通过卷挂载从 `data/clusters/` 映射
- **macOS**：DST 服务器文件从主机 `data/dst_server/` 挂载
- **Linux**：DST 服务器通过 SteamCMD 安装在容器内部
