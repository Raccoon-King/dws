package scanner

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ExtractText extracts text from various file formats
func ExtractText(data []byte, filename string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".pdf":
		return extractPDFText(data)
	case ".txt":
		return string(data), nil
	case ".html", ".htm":
		return extractHTMLText(data)
	case "":
		// No extension - try to extract as text
		return string(data), nil
	default:
		// Check if it looks like binary data
		if isBinaryData(data) {
			return "", fmt.Errorf("unsupported file format: %s", ext)
		}
		// Try to extract as plain text for unknown text-like formats
		return string(data), nil
	}
}

// isBinaryData performs a basic check to see if data is likely binary
func isBinaryData(data []byte) bool {
	// Check first 512 bytes for null bytes which are common in binary files
	for i, b := range data {
		if i > 512 {
			break
		}
		if b == 0 {
			return true
		}
		if i >= 3 {
			// Check for PNG signature (89 PNG)
			if len(data) >= 4 {
				prefix := data[:4]
				if prefix[0] == 0x89 && prefix[1] == 0x50 && prefix[2] == 0x4E && prefix[3] == 0x47 {
					return true
				}
			}
		}
		if i >= 1 {
			// Check for BMP signature (BM)
			if len(data) >= 2 {
				prefix := data[:2]
				if prefix[0] == 0x42 && prefix[1] == 0x4D {
					return true
				}
			}
		}
	}
	return false
}

// extractPDFText extracts text from PDF files
func extractPDFText(data []byte) (string, error) {
	// TODO: Implement PDF text extraction
	// For now, return a placeholder
	return "PDF text extraction not implemented", nil
}

// extractHTMLText extracts text from HTML files
func extractHTMLText(data []byte) (string, error) {
	html := string(data)

	// Simple HTML-to-text conversion (very basic)
	// Remove HTML tags
	text := strings.ReplaceAll(html, "<script>", "")
	text = strings.ReplaceAll(text, "</script>", "")
	text = strings.ReplaceAll(text, "<style>", "")
	text = strings.ReplaceAll(text, "</style>", "")

	// Basic tag removal
	start := strings.Index(text, "<")
	for start != -1 {
		end := strings.Index(text[start:], ">")
		if end == -1 {
			break
		}
		text = text[:start] + text[start+end+1:]
		start = strings.Index(text, "<")
	}

	// TODO: Decode HTML entities later if needed
	// Currently using basic extraction

	return strings.TrimSpace(text), nil
}
