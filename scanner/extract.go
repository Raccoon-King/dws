package scanner

import (
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
	default:
		// Try to extract as plain text for unknown formats
		return string(data), nil
	}
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
