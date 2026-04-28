package main

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ledongthuc/pdf"
	"github.com/xuri/excelize/v2"
)

const maxChars = 4_000_000

var plainTextExts = map[string]bool{
	".txt": true, ".md": true, ".csv": true, ".log": true,
	".json": true, ".yaml": true, ".yml": true, ".toml": true,
	".xml": true, ".html": true, ".htm": true, ".js": true,
	".ts": true, ".go": true, ".py": true, ".rs": true,
	".java": true, ".c": true, ".cpp": true, ".h": true,
	".sh": true, ".bash": true, ".fish": true, ".zsh": true,
	".sql": true, ".graphql": true, ".tf": true, ".env": true,
}

func ExtractText(data []byte, filename string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filename))

	if plainTextExts[ext] {
		return truncate(string(data)), nil
	}

	switch ext {
	case ".pdf":
		return extractPDF(data)
	case ".docx":
		return extractDOCX(data)
	case ".xlsx":
		return extractXLSX(data)
	case ".doc", ".xls":
		return "", fmt.Errorf("legacy Office format (%s) not supported; convert to .docx/.xlsx", ext)
	default:
		if looksLikeText(data) {
			return truncate(string(data)), nil
		}
		return "", fmt.Errorf("unsupported or binary file type: %s", ext)
	}
}

func truncate(s string) string {
	if len(s) <= maxChars {
		return s
	}
	return s[:maxChars] + "\n\n[... truncated at 4M characters ...]"
}

func looksLikeText(data []byte) bool {
	if len(data) == 0 {
		return true
	}
	printable := 0
	check := data
	if len(check) > 512 {
		check = check[:512]
	}
	for _, b := range check {
		if b >= 32 || b == '\n' || b == '\r' || b == '\t' {
			printable++
		}
	}
	return float64(printable)/float64(len(check)) > 0.8
}

func extractPDF(data []byte) (string, error) {
	r, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("failed to parse PDF: %w", err)
	}

	var sb strings.Builder
	for i := 1; i <= r.NumPage(); i++ {
		page := r.Page(i)
		if page.V.IsNull() {
			continue
		}
		text, err := page.GetPlainText(nil)
		if err != nil {
			continue
		}
		sb.WriteString(text)
		if sb.Len() >= maxChars {
			break
		}
	}

	return truncate(sb.String()), nil
}

func extractDOCX(data []byte) (string, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("failed to open DOCX: %w", err)
	}

	for _, f := range r.File {
		if f.Name != "word/document.xml" {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return "", fmt.Errorf("failed to open document.xml: %w", err)
		}
		defer rc.Close()

		var buf bytes.Buffer
		if _, err := buf.ReadFrom(rc); err != nil {
			return "", err
		}

		return truncate(xmlToText(buf.Bytes())), nil
	}

	return "", fmt.Errorf("document.xml not found in DOCX")
}

func xmlToText(data []byte) string {
	dec := xml.NewDecoder(bytes.NewReader(data))
	var sb strings.Builder

	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "p" {
				sb.WriteByte('\n')
			}
		case xml.CharData:
			sb.Write([]byte(t))
		}
	}

	return strings.TrimSpace(sb.String())
}

func extractXLSX(data []byte) (string, error) {
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("failed to open XLSX: %w", err)
	}
	defer f.Close()

	var sb strings.Builder
	for _, sheet := range f.GetSheetList() {
		rows, err := f.GetRows(sheet)
		if err != nil {
			continue
		}
		sb.WriteString("Sheet: ")
		sb.WriteString(sheet)
		sb.WriteByte('\n')

		for _, row := range rows {
			sb.WriteString(strings.Join(row, "\t"))
			sb.WriteByte('\n')
		}
		sb.WriteByte('\n')

		if sb.Len() >= maxChars {
			break
		}
	}

	return truncate(sb.String()), nil
}
