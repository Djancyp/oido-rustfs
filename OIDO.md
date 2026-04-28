# Oido RustFS Extension

Read files from RustFS object storage and inject their text content into the LLM context.

## Available Tools

### `rustfs_read_file`
Fetch a file from RustFS by storage key and return its extracted text content.

**Parameters:**
- `key` (string, required): Storage key of the file (e.g. `main.go`, `report.pdf`)
- `bucket` (string, optional): Bucket to read from. Defaults to first bucket in `RUSTFS_BUCKET`.

**Supported formats:** txt, md, csv, pdf, docx, xlsx, json, yaml, toml, xml, html, and most plain-text source files.

**Returns:** Full text content of the file, truncated at 4M characters (~1M tokens) if too large.

## Example Usage

```
User: Attached files: - main.go (storage key: main.go)

A: I'll read that file from storage.

[Uses rustfs_read_file with key="main.go"]

File: main.go (bucket: chat-attachments)

package main
...
```

## When to Use

- User attaches a file reference with a storage key
- User asks to read or analyze a file stored in RustFS
- User wants to review document content (PDF, DOCX, XLSX)

## Notes

- **Default bucket**: `chat-attachments` (first value in `RUSTFS_BUCKET`)
- **Multiple buckets**: Set `RUSTFS_BUCKET=bucket1,bucket2` — first is default, all are allowed
- **Large files**: Content truncated at 4M characters with a notice
- **Binary formats**: `.doc` and `.xls` (legacy Office) not supported — convert to `.docx`/`.xlsx`

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `RUSTFS_BASE_URL` | RustFS server URL (e.g. `http://localhost:9000`) | *(required)* |
| `RUSTFS_ACCESS_KEY` | Access key | *(required)* |
| `RUSTFS_SECRET_KEY` | Secret key | *(required)* |
| `RUSTFS_BUCKET` | Allowed buckets, comma-separated | `chat-attachments` |
