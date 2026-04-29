package main

import (
	"context"
	"fmt"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type MCPHandler struct {
	storage *StorageClient
	cfg     *Config
}

type ReadFileArgs struct {
	Key    string `json:"key" jsonschema:"Storage key of the file to read (e.g. 'main.go')"`
	Bucket string `json:"bucket,omitempty" jsonschema:"Bucket name (optional, overrides default bucket)"`
}

type UploadFileArgs struct {
	Key     string `json:"key" jsonschema:"Storage key to save the file as (e.g. 'notes.md')"`
	Content string `json:"content" jsonschema:"Text content to upload"`
	Bucket  string `json:"bucket,omitempty" jsonschema:"Bucket name (optional, uses default bucket)"`
}

func RunMCPServer() {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("Config error: %v", err)
	}

	storage, err := NewStorageClient(cfg)
	if err != nil {
		log.Fatalf("Storage client error: %v", err)
	}

	handler := &MCPHandler{storage: storage, cfg: cfg}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "oido-rustfs",
		Version: "1.0.0",
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name: "rustfs_upload_file",
		Description: `Upload text content to RustFS object storage and save it under the given key.
Use this tool when the user wants to save, store, export, or write content to a file.
The content type is inferred from the file extension (e.g. .md → markdown, .csv → CSV, .json → JSON).
Supported formats: txt, md, csv, json, yaml, toml, xml, html, js, ts, go, py, rs, sql, and most plain-text source files.
Parameters:
  - key (required): filename/path to store as, e.g. "notes.md" or "output/report.txt"
  - content (required): full text content to save
  - bucket (optional): target bucket; defaults to the first configured bucket`,
	}, handler.HandleUploadFile)

	mcp.AddTool(server, &mcp.Tool{
		Name: "rustfs_read_file",
		Description: `Read a file from RustFS object storage and return its full text content for use in the conversation.
AUTO-TRIGGER: call this tool automatically whenever the user message contains an attached file reference in the format "Attached files: - <filename> (storage key: <key>)" — extract the storage key and use it as the key parameter without asking the user.
Supports automatic text extraction for: txt, md, csv, json, yaml, toml, xml, html, source code files, PDF, DOCX, and XLSX.
Content is truncated at ~1M tokens if the file is very large.
Parameters:
  - key (required): storage key of the file, e.g. "main.go" or "reports/q1.pdf"
  - bucket (optional): bucket to read from; if omitted, searches all configured buckets in order and returns the first match`,
	}, handler.HandleReadFile)

	ctx := context.Background()
	log.Println("Oido RustFS MCP Server starting on stdio...")

	if err := server.Run(ctx, mcp.NewStdioTransport()); err != nil {
		log.Fatalf("MCP server error: %v", err)
	}
}

func (h *MCPHandler) HandleReadFile(ctx context.Context, _ *mcp.ServerSession, params *mcp.CallToolParamsFor[ReadFileArgs]) (*mcp.CallToolResult, error) {
	args := params.Arguments

	if args.Key == "" {
		return errResult("key parameter is required"), nil
	}

	// If bucket explicitly specified, try that one only.
	buckets := h.cfg.Buckets
	if args.Bucket != "" {
		buckets = []string{args.Bucket}
	}

	var data []byte
	var foundBucket string
	for _, bucket := range buckets {
		d, err := h.storage.GetObject(ctx, bucket, args.Key)
		if err == nil {
			data = d
			foundBucket = bucket
			break
		}
	}

	if foundBucket == "" {
		return errResult(fmt.Sprintf("%q not found in any bucket: %v", args.Key, h.cfg.Buckets)), nil
	}

	text, err := ExtractText(data, args.Key)
	if err != nil {
		return errResult(fmt.Sprintf("failed to extract text from %s: %v", args.Key, err)), nil
	}

	result := fmt.Sprintf("File: %s (bucket: %s)\n\n%s", args.Key, foundBucket, text)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result},
		},
	}, nil
}

func (h *MCPHandler) HandleUploadFile(ctx context.Context, _ *mcp.ServerSession, params *mcp.CallToolParamsFor[UploadFileArgs]) (*mcp.CallToolResult, error) {
	args := params.Arguments

	if args.Key == "" {
		return errResult("key parameter is required"), nil
	}
	if args.Content == "" {
		return errResult("content parameter is required"), nil
	}

	bucket := args.Bucket
	if bucket == "" {
		bucket = h.cfg.DefaultBucket
	}

	if !h.cfg.IsAllowedBucket(bucket) {
		return errResult(fmt.Sprintf("bucket %q not in allowed list: %v", bucket, h.cfg.Buckets)), nil
	}

	if err := h.storage.PutObject(ctx, bucket, args.Key, args.Content); err != nil {
		return errResult(fmt.Sprintf("failed to upload %s/%s: %v", bucket, args.Key, err)), nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Uploaded %s to bucket %s (%d bytes)", args.Key, bucket, len(args.Content))},
		},
	}, nil
}

func errResult(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: "Error: " + msg},
		},
		IsError: true,
	}
}
