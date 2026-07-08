<div align="center">

# tdl-filegram

### A Web visual management UI for [tdl](https://github.com/iyear/tdl) — paste a `t.me` link and videos, images, and files land on your disk.

Built on top of the tdl core download engine, with a Web UI crafted in Vue 3 + Ant Design Vue. No more command line — QR login, paste a link, real-time progress, and in-browser preview, all done in your browser. The frontend is embedded into a single binary via `go:embed` — **zero external dependencies to deploy**.

[![Stars](https://img.shields.io/github/stars/weilaifeng/tdl-filegram?style=flat-square&logo=github&color=yellow)](https://github.com/weilaifeng/tdl-filegram/stargazers)
[![Forks](https://img.shields.io/github/forks/weilaifeng/tdl-filegram?style=flat-square&logo=github&color=blue)](https://github.com/weilaifeng/tdl-filegram/network/members)
[![Issues](https://img.shields.io/github/issues/weilaifeng/tdl-filegram?style=flat-square&logo=github)](https://github.com/weilaifeng/tdl-filegram/issues)
[![License](https://img.shields.io/github/license/weilaifeng/tdl-filegram?style=flat-square&color=green)](./LICENSE)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev)
[![Vue](https://img.shields.io/badge/Vue-3-4FC08D?style=flat-square&logo=vue.js&logoColor=white)](https://vuejs.org)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat-square&logo=docker&logoColor=white)](./Dockerfile)

[💡 Highlights](#-highlights) · [📥 Quick Start](#-quick-start) · [🎬 Usage](#-usage) · [🌐 Tech Stack](#-tech-stack) · [📊 API](#-api) · [⚠️ Risk Notice](#-risk-notice--disclaimer) · [❓ FAQ](#-faq) · [🗺️ Roadmap](#-roadmap)

[中文](./README.md) · 🌐 **English**

</div>

---

## 💡 Highlights

The biggest highlight of this project: **a ready-to-use Web visual management UI for [tdl](https://github.com/iyear/tdl), the powerful Telegram command-line downloader.**

tdl is an excellent CLI tool, but it comes with many command-line flags, terminal-based login interaction, and no persistent view of task history or real-time progress. tdl-filegram wraps a Web UI around tdl without touching its core capabilities:

| Aspect | tdl (CLI) | tdl-filegram (this project) |
| --- | --- | --- |
| Interaction | CLI flags + config files | Browser GUI, point and click |
| Login | Terminal phone number / code input | Phone QR login, with 2FA support |
| Task progress | Terminal text output, gone when closed | Real-time progress bar + task list, history kept |
| File preview | Needs a local player | In-browser playback / viewing |
| Deployment | Single binary | Single binary (frontend go:embed) / Docker |

Under the hood it reuses tdl's core modules (`core/downloader` multi-threaded engine, `core/tmedia` media parsing, `core/tclient` client wrapper, `core/storage` session storage, `core/dcpool` connection pool), and layers an HTTP API plus a Web frontend on top — presenting tdl's download power in a more intuitive, user-friendly way. **CLI capabilities, GUI experience.**

## ✨ Features

- **QR Login** — Scan with your phone, supports 2FA. Sessions persist in BoltDB, no re-login after restart
- **Multi-threaded Download** — Powered by the [tdl](https://github.com/iyear/tdl) engine, configurable thread count and concurrency
- **Task Management** — Create / query / list / track progress with pagination and real-time download progress
- **Online Preview** — Preview downloaded files directly in the browser (video playback, image viewing)
- **Proxy Support** — http / https / socks5 proxy for network-restricted environments
- **Single-file Deployment** — Frontend embedded in binary, just one executable + SQLite, CGO-free
- **Docker Deployment** — Multi-stage build, env-var configuration, `docker compose up` ready to go

## 🎯 Who Is It For

| You Are | What You Can Do |
| --- | --- |
| **Telegram channel/group archiver** | Paste message links to batch-download channel videos, images, and files locally |
| **User in a restricted network** | Built-in http / socks5 proxy support keeps you connected to Telegram |
| **Want in-browser preview** | Play downloaded videos / view images right in the browser — no local player needed |
| **NAS / home server owner** | Single-container Docker deploy, mount the download dir, run it long-term as a Telegram offline downloader |
| **Don't want a pile of dependencies** | Frontend `go:embed`-ed into the binary, pure-Go SQLite driver (no CGO), one file does it all |
| **Find the tdl CLI tedious** | Want tdl's download power without memorizing CLI flags — Web UI with QR login, paste a link, and download |

## ⚡ Quick Start

### Option 1: Pull the Image (Recommended)

Images are published to GitHub Container Registry. Pick the tag matching your CPU architecture — no local build needed:

```bash
# ARM64 (Apple Silicon / Raspberry Pi, etc.)
docker pull ghcr.io/studynoweekend/tdl-filegram:1.0-arm64

# AMD64 (Intel / AMD)
docker pull ghcr.io/studynoweekend/tdl-filegram:1.0-amd64
```

Start the container (amd64 shown below; for arm64 swap the tag to `1.0-arm64`):

```bash
docker run -d --name tdl-filegram \
  --restart unless-stopped \
  -p 8744:8744 \
  -p 8743:8743 \
  -v ./data:/data \
  -v ./downloads:/downloads \
  -e DB_PATH=/data/tdl-filegram.db \
  -e DOWNLOAD_DIR=/downloads \
  -e DOWNLOAD_THREADS=4 \
  -e DOWNLOAD_LIMIT=2 \
  -e TG_APP_ID=your_app_id \
  -e TG_APP_HASH=your_app_hash \
  -e TG_DATA_DIR=/data/.tdl \
  -e TG_NAMESPACE=default \
  -e TG_POOL_SIZE=8 \
  -e TG_RECONNECT_TIMEOUT=5m \
  -e TG_PROXY=http://192.168.1.100:7890 \
  ghcr.io/studynoweekend/tdl-filegram:1.0-amd64
```

### Option 2: Build the Image Locally

```bash
# Build the image
docker build -t tdl-filegram .

# Start the container
docker run -d --name tdl-filegram \
  --restart unless-stopped \
  -p 8744:8744 \
  -p 8743:8743 \
  -v ./data:/data \
  -v ./downloads:/downloads \
  -e DB_PATH=/data/tdl-filegram.db \
  -e DOWNLOAD_DIR=/downloads \
  -e DOWNLOAD_THREADS=4 \
  -e DOWNLOAD_LIMIT=2 \
  -e TG_APP_ID=your_app_id \
  -e TG_APP_HASH=your_app_hash \
  -e TG_DATA_DIR=/data/.tdl \
  -e TG_NAMESPACE=default \
  -e TG_POOL_SIZE=8 \
  -e TG_RECONNECT_TIMEOUT=5m \
  -e TG_PROXY=http://192.168.1.100:7890 \
  tdl-filegram
```

### Option 3: Local Development

**Prerequisites:** Go 1.25+, Node.js 18+

```bash
# Build the frontend
cd web && npm install && npm run build && cd ..

# Fetch deps and run the backend
go mod tidy && go run cmd/api/main.go
```

For dev mode, start the frontend dev server and backend separately:

```bash
# Terminal 1: backend (listens on :8743)
go mod tidy && go run cmd/api/main.go

# Terminal 2: frontend dev server (listens on :5173, auto-proxies /api → :8743)
cd web && npm run dev
# Open http://localhost:5173
```

**Port Reference:**

| Port | Service | Description |
| --- | --- | --- |
| `8744` | nginx (frontend) | Frontend pages + reverse proxy `/api` to backend |
| `8743` | Go backend | API service, also directly accessible |

- Browser frontend: `http://localhost:8744`
- Direct API: `http://localhost:8743`

## 🎬 Usage

1. Start the service and open `http://localhost:8744` in your browser
2. Click "Scan to Login" and scan the QR code with your Telegram mobile app
3. If 2FA is enabled, enter your password
4. Once logged in, paste a `t.me` message link and click download
5. The task list shows real-time progress; preview or download the file once complete

## 🌐 Tech Stack

| Layer | Technology | Description |
| --- | --- | --- |
| Web Framework | [Gin](https://github.com/gin-gonic/gin) | HTTP routing & middleware |
| Telegram MTProto | [gotd/td](https://github.com/gotd/td) + [tdl core](https://github.com/iyear/tdl) | Telegram communication + download engine |
| Database | SQLite ([glebarez/sqlite](https://github.com/glebarez/sqlite)) | Pure-Go driver, CGO-free |
| Logging | [zap](https://github.com/uber-go/zap) + lumberjack | Structured logging + rolling archives |
| Frontend | Vue 3 + Ant Design Vue 4 + Vite | Embedded into binary (go:embed) |
| Session Storage | BoltDB ([bbolt](https://github.com/etcd-io/bbolt)) | Telegram session persistence |

## 🔧 Configuration

Config priority: **Environment variables > config.yaml > code defaults**.

<details>
<summary><b>Config file (config/config.yaml)</b></summary>

```yaml
app:
  name: tdl-filegram
  port: "8743"
  env: debug              # debug / release

log:
  level: info             # info / debug
  dir: logs

database:
  path: data/tdl-filegram.db # SQLite database file path

download:
  dir: downloads          # download output directory
  threads: 4              # threads per file
  limit: 2                # max concurrent tasks

telegram:
  app_id:             # Telegram API ID
  app_hash: ""
  data_dir: .tdl          # BoltDB session storage dir
  namespace: default      # session namespace
  pool_size: 8            # DC connection pool size
  reconnect_timeout: 5m   # reconnect backoff timeout
  proxy: ""               # proxy address, e.g. http://127.0.0.1:7890
```

</details>

<details>
<summary><b>Environment variables (for Docker)</b></summary>

All config items can be overridden via environment variables:

| Env Variable | Config Key | Docker Default |
| --- | --- | --- |
| `APP_PORT` | `app.port` | `8743` |
| `DB_PATH` | `database.path` | `/data/tdl-filegram.db` |
| `DOWNLOAD_DIR` | `download.dir` | `/downloads` |
| `DOWNLOAD_THREADS` | `download.threads` | `4` |
| `DOWNLOAD_LIMIT` | `download.limit` | `2` |
| `TG_APP_ID` | `telegram.app_id` | `` |
| `TG_APP_HASH` | `telegram.app_hash` | `` |
| `TG_DATA_DIR` | `telegram.data_dir` | `/data/.tdl` |
| `TG_NAMESPACE` | `telegram.namespace` | `default` |
| `TG_POOL_SIZE` | `telegram.pool_size` | `8` |
| `TG_RECONNECT_TIMEOUT` | `telegram.reconnect_timeout` | `5m` |
| `TG_PROXY` | `telegram.proxy` | empty (no proxy) |

</details>

<details>
<summary><b>Docker volumes</b></summary>

| Container Path | Purpose |
| --- | --- |
| `/data` | SQLite database + Telegram session persistence |
| `/downloads` | Downloaded media file output directory |

</details>

<details>
<summary><b>How to get Telegram API credentials?</b></summary>

`app_id` / `app_hash` are the Telegram MTProto API application credentials:

1. **Apply yourself (recommended)** — Log in to [my.telegram.org](https://my.telegram.org) → "API development tools", fill in app info to get your own credentials.
2. **Use public credentials** — You may also use the public Telegram Desktop credentials (`app_id: 2040`) as a temporary solution, but this is equivalent to impersonating the official desktop client. Heavy use carries a risk-control risk; replace with your own for production.

</details>

## 📊 API

All endpoints return a unified structure:

```json
{
  "code": 0,
  "msg": "success",
  "data": {},
  "trace_id": "xxx"
}
```

<details>
<summary><b>Login endpoints</b></summary>

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/api/login/status` | Query login status (ready, logged in, QR URL, 2FA status) |
| `POST` | `/api/login/qr/start` | Start QR login flow |
| `POST` | `/api/login/qr/2fa` | Submit 2FA password |

**`login_status` values:**

| Value | Meaning |
| --- | --- |
| `""` (empty) | Idle, no login initiated |
| `pending` | QR code generated, waiting for scan confirmation |
| `need_2fa` | 2FA password required |
| `success` | Login successful |
| `error` | Login failed |

</details>

<details>
<summary><b>Download endpoints</b></summary>

| Method | Path | Description |
| --- | --- | --- |
| `POST` | `/api/download` | Create a download task |

Request body:

```json
{
  "url": "https://t.me/channel/123"
}
```

Response:

```json
{
  "job_id": "uuid"
}
```

</details>

<details>
<summary><b>Job endpoints</b></summary>

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/api/jobs?page=1&page_size=20` | Paginated task list |
| `GET` | `/api/jobs/:id` | Single task detail (with real-time progress) |
| `GET` | `/api/jobs/:id/file` | Download / preview a completed task's file |

**`status` values:** `pending` / `downloading` / `success` / `failed`

</details>

<details>
<summary><b>Other</b></summary>

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/health` | Health check |

</details>

## 📁 Project Structure

```
tdl-filegram/
├── cmd/api/              # Entry point
├── bootstrap/            # Startup init (config, log, database)
├── config/               # Config files
├── enum/                 # Error codes & constants
├── internal/
│   ├── controller/       # HTTP controllers
│   ├── dto/              # Request / response structs
│   ├── logic/            # Business logic
│   ├── middleware/       # Gin middleware (CORS, Recovery, Trace)
│   ├── model/            # Data models & GORM operations
│   └── router/           # Route registration
├── pkg/telegram/         # Telegram engine (login, download, storage)
├── utils/                # Utility functions
├── web/                  # Frontend source + embedded build
├── Dockerfile            # Multi-stage build
├── docker-compose.yml    # Docker Compose
└── Makefile              # Build/run shortcuts
```

## 🛡️ Notes

- Configure a proxy on first launch if your network can't reach Telegram directly
- Sessions persist in `data_dir` (BoltDB) — no re-login after restart
- Reusing Telegram Desktop credentials means impersonating a client; heavy use carries a ban risk. Replace with your own API credentials for production
- Downloaded files are saved in the directory configured by `download.dir`

## ⚠️ Risk Notice & Disclaimer

**Using this project may cause your Telegram account to be restricted or banned. By using it, you acknowledge and voluntarily assume all risks. The author is not responsible for any account loss, data loss, or other consequences.**

This project does not interact with Telegram in the officially recommended way. The main points:

1. **Unofficial SDK** — Telegram officially provides only TDLib (C++) and Bot API (HTTP), with no Go SDK. This project uses [gotd/td](https://github.com/gotd/td), a community third-party implementation of the MTProto protocol, rather than the official TDLib. The library's own author reminds users to read the "How To Not Get Banned" guide.
2. **User account login** — This project logs in as a user account (not a Bot) so it can download media from protected channels — something the official Bot API cannot do. This also means you act as a "third-party user client", which is more likely to trigger risk control.
3. **Credential impersonation** — Reusing the public Telegram Desktop credentials (`app_id: 2040`) is equivalent to impersonating the official desktop client to the server. Heavy use significantly raises the ban risk; replace with your own credentials.

All of the above may trigger Telegram's risk control mechanisms. Please fully understand the risks before use.

### How to reduce the risk of being banned?

- Log in using an official client session
- Stick to the default download and upload options as much as possible — **do not set overly large `threads` and `size`**
- Do not log in with the same account on multiple devices at the same time
- Do not download or upload too many files simultaneously
- Become a Telegram Premium member 😅

## ❓ FAQ

<details>
<summary><b>Do I need to re-scan after restarting the service?</b></summary>

No. Telegram sessions persist in `data_dir` (BoltDB). As long as the `/data` volume is mounted, the login state is restored automatically after restart.

</details>

<details>
<summary><b>What if my network can't reach Telegram?</b></summary>

Set the `TG_PROXY` environment variable. It supports http / https / socks5 proxies, e.g. `http://127.0.0.1:7890` or `socks5://127.0.0.1:1080`.

</details>

<details>
<summary><b>Where are downloaded files stored?</b></summary>

In the directory configured by `download.dir` (Docker default: `/downloads`). Mount it to the host to access files directly.

</details>

<details>
<summary><b>What types of content can be downloaded?</b></summary>

Videos, images, and files from `t.me` message links, powered by the tdl multi-threaded download engine.

</details>

## 🗺️ Roadmap

- ✅ QR login + 2FA
- ✅ Multi-threaded download + task management + online preview
- ✅ Docker deployment + proxy support
- 🔲 Folder / channel batch download
- 🔲 Download speed limit
- 🔲 More file format previews

## 💌 Acknowledgments

This project stands on the shoulders of giants. Special thanks to:

- **[tdl](https://github.com/iyear/tdl)** — by [@iyear](https://github.com/iyear), the Telegram download powerhouse. tdl-filegram reuses tdl's core modules (`core/downloader` multi-threaded engine, `core/tmedia` media parsing, `core/tclient` client wrapper, `core/storage` session storage, `core/dcpool` connection pool). Without tdl, this project wouldn't exist.
- **[gotd/td](https://github.com/gotd/td)** — A complete Telegram MTProto API implementation in Go. All underlying Telegram communication in this project is built on it.

## ⭐ Star history

[![Stargazers over time](https://api.star-history.com/svg?repos=weilaifeng/tdl-filegram&type=Date)](https://star-history.com/#weilaifeng/tdl-filegram&Date)

## 📜 License

[AGPL-3.0](./LICENSE) — This project reuses the core modules of [tdl](https://github.com/iyear/tdl) (AGPL-3.0). As required by its license, this project is also released under AGPL-3.0. Any derivative work based on this project must be open-sourced under the same license.

---

<div align="center">

**Found it useful? A ⭐ means a lot to the author.**

[⬆ Back to top](#tdl-filegram) · [📥 Quick Start](#-quick-start) · [💬 Open an Issue](https://github.com/weilaifeng/tdl-filegram/issues)

</div>
