# Oido RustFS MCP Extension

Read files from RustFS object storage and inject their text content into the LLM context.

## Features

- **File Reading**: Fetch any file from RustFS by storage key
- **Text Extraction**: Automatic extraction for PDF, DOCX, XLSX, and plain-text formats
- **Multi-bucket**: Support multiple buckets with a configurable default
- **Large File Handling**: Truncates at 4M characters (~1M tokens) with a notice

## Installation

### Option 1: Upload via Plugins UI (Recommended)

1. Download the latest release zip for your platform from [GitHub Releases](../../releases)
   - Linux: `oido-rustfs-linux-amd64.zip`
   - macOS (Apple Silicon): `oido-rustfs-darwin-arm64.zip`
2. Open Oido Studio → Plugins UI
3. Upload the zip file
4. Configure settings (base URL, access key, secret key, bucket) in the plugin settings panel

### Option 2: Build from Source

```bash
git clone <repo-url>
cd oido-rustfs
make build
```

### Option 3: Manual Install from Release Artifacts

```bash
curl -LO https://github.com/<owner>/<repo>/releases/latest/download/oido-rustfs-linux-amd64.zip
unzip oido-rustfs-linux-amd64.zip -d oido-rustfs
./oido-rustfs/oido-rustfs-mcp
```

## Requirements

- Go 1.23+
- RustFS (or any S3-compatible object storage)

## Setup

### 1. Start RustFS

Follow the [RustFS documentation](https://rustfs.com) to start a server, or point to any S3-compatible endpoint (MinIO, AWS S3, etc.).

### 2. Configure Extension

Set the following environment variables (or configure via plugin settings):

| Variable | Description | Default |
|----------|-------------|---------|
| `RUSTFS_BASE_URL` | Server URL (e.g. `http://localhost:9000`) | *(required)* |
| `RUSTFS_ACCESS_KEY` | Access key | *(required)* |
| `RUSTFS_SECRET_KEY` | Secret key | *(required)* |
| `RUSTFS_BUCKET` | Allowed buckets, comma-separated. First = default. | `chat-attachments` |

### 3. Multiple Buckets

```bash
RUSTFS_BUCKET=chat-attachments,uploads,documents
```

The first bucket is the default. Pass `bucket` in the tool call to override.

## Build

```bash
make build
```

## Package for Distribution

```bash
make dist
```

Creates `dist/oido-rustfs.zip` for upload via the Plugins UI.

## Tool

### `rustfs_read_file`

Fetch a file from RustFS by storage key and return its extracted text content.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `key` | string | yes | Storage key (e.g. `main.go`, `report.pdf`) |
| `bucket` | string | no | Bucket override (must be in allowed list) |

**Supported formats:**

| Format | Extensions |
|--------|-----------|
| Plain text | txt, md, csv, log, json, yaml, toml, xml, html, and most source files |
| PDF | pdf |
| Word | docx |
| Excel | xlsx |

> **Note:** Legacy `.doc` and `.xls` formats are not supported. Convert to `.docx`/`.xlsx` first.

## Architecture

```
┌─────────────────┐    stdio     ┌───────────────────┐
│  Oido Studio    │ ◄──────────► │  oido-rustfs-mcp  │
│                 │              │                   │
│                 │              │  ┌─────────────┐  │
│                 │              │  │ S3 Client   │  │──► RustFS
│                 │              │  └─────────────┘  │
│                 │              │  ┌─────────────┐  │
│                 │              │  │ Extractor   │  │
│                 │              │  │ pdf/docx/   │  │
│                 │              │  │ xlsx/text   │  │
│                 │              │  └─────────────┘  │
└─────────────────┘              └───────────────────┘
```

## License

MIT
