# Oido RustFS Extension

Read files from RustFS object storage and inject their text content into the LLM context.

## Available Tools

### `rustfs_read_file`
Fetch a file from RustFS by storage key and return its extracted text content.

**Parameters:**
- `key` (string, required): Storage key of the file (e.g. `main.go`, `report.pdf`)
- `bucket` (string, optional): Bucket to read from. Defaults to first bucket in `OIDO_RUSTFS_BUCKET`.

**Supported formats:** txt, md, csv, pdf, docx, xlsx, json, yaml, toml, xml, html, and most plain-text source files.

**Returns:** Full text content of the file, truncated at 4M characters (~1M tokens) if too large.

### `rustfs_search_files`
Search for files by key pattern and optionally by content.

**Parameters:**
- `pattern` (string, required): Glob pattern or prefix. Wildcards `*`, `?`, `[...]` supported. No wildcards = prefix match (all keys starting with string). Examples: `"*.md"`, `"reports/*"`, `"budget"`, `"2024/report*"`.
- `query` (string, optional): Case-insensitive content search. If set, only files containing this text are returned.
- `bucket` (string, optional): Restrict search to one bucket. Default: all configured buckets.
- `max_results` (int, optional): Max results (default 20, max 50).

**Returns:** Full text of each matching file.

**Best practices:**
- Use specific patterns to avoid listing too many objects
- Narrow by bucket when possible
- Query searches within extracted text of matching files

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

- **Default bucket**: `chat-attachments` (first value in `OIDO_RUSTFS_BUCKET`)
- **Multiple buckets**: Set `OIDO_RUSTFS_BUCKET=bucket1,bucket2` — first is default, all are allowed
- **Large files**: Content truncated at 4M characters with a notice
- **Binary formats**: `.doc` and `.xls` (legacy Office) not supported — convert to `.docx`/`.xlsx`

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `OIDO_RUSTFS_BASE_URL` | RustFS server URL (e.g. `http://localhost:9000`) | *(required)* |
| `OIDO_RUSTFS_ACCESS_KEY` | Access key | *(required)* |
| `OIDO_RUSTFS_SECRET_KEY` | Secret key | *(required)* |
| `OIDO_RUSTFS_BUCKET` | Allowed buckets, comma-separated | `chat-attachments` |
