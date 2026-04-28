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
		Name:        "rustfs_read_file",
		Description: "Read a file from RustFS object storage and return its text content. Supports txt, md, csv, pdf, docx, xlsx, and most plain-text formats.",
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

	bucket := args.Bucket
	if bucket == "" {
		bucket = h.cfg.DefaultBucket
	}

	if !h.cfg.IsAllowedBucket(bucket) {
		return errResult(fmt.Sprintf("bucket %q not in allowed list: %v", bucket, h.cfg.Buckets)), nil
	}

	data, err := h.storage.GetObject(ctx, bucket, args.Key)
	if err != nil {
		return errResult(fmt.Sprintf("failed to fetch %s/%s: %v", bucket, args.Key, err)), nil
	}

	text, err := ExtractText(data, args.Key)
	if err != nil {
		return errResult(fmt.Sprintf("failed to extract text from %s: %v", args.Key, err)), nil
	}

	result := fmt.Sprintf("File: %s (bucket: %s)\n\n%s", args.Key, bucket, text)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result},
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
