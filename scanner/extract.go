package scanner

import (
	"bytes"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dslipak/pdf"
)

// ExtractText converts various document formats into plain text.
func ExtractText(data []byte, filename string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	var extractedText string
	switch ext {
	case ".pdf":
		reader := bytes.NewReader(data)
		pdfReader, err := pdf.NewReader(reader, int64(len(data)))
		if err != nil {
			return "", fmt.Errorf("failed to create PDF reader: %w", err)
		}

		var sb strings.Builder
		for i := 1; i <= pdfReader.NumPage(); i++ {
			page := pdfReader.Page(i)
			if page.V.IsNull() {
				continue
			}
			text, err := page.GetPlainText(nil)
			if err != nil {
				return "", fmt.Errorf("failed to get plain text from page %d: %w", i, err)
			}
			sb.WriteString(text)
		}
		extractedText = sb.String()
	case ".html", ".htm":
		// First, remove <script> and <style> blocks
		re := regexp.MustCompile("(?s)<(script|style)>.*?</(script|style)>")
		cleaned := re.ReplaceAllString(string(data), " ")
		// Then, remove all other tags
		re = regexp.MustCompile("<[^>]+>")
		extractedText = re.ReplaceAllString(cleaned, " ")
		case ".yaml", ".yml", ".txt", ".json", ".xml":
		extractedText = string(data)
	default:
		return "", fmt.Errorf("unsupported file type: %s", ext)
	}

	return extractedText, nil
}
