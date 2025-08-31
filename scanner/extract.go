package scanner

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
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
	case ".docx":
		readerAt := bytes.NewReader(data)
		zr, err := zip.NewReader(readerAt, int64(len(data)))
		if err != nil {
			return "", fmt.Errorf("failed to create zip reader: %w", err)
		}
		var docFile *zip.File
		for _, f := range zr.File {
			if f.Name == "word/document.xml" {
				docFile = f
				break
			}
		}
		if docFile == nil {
			return "", fmt.Errorf("document.xml not found in docx")
		}
		rc, err := docFile.Open()
		if err != nil {
			return "", fmt.Errorf("failed to open document.xml: %w", err)
		}
		defer rc.Close()
		xmlData, err := io.ReadAll(rc)
		if err != nil {
			return "", fmt.Errorf("failed to read document.xml: %w", err)
		}
		re := regexp.MustCompile("<[^>]+>")
		extractedText = re.ReplaceAllString(string(xmlData), " ")
	case ".rtf":
		// Strip RTF control words and braces to approximate plain text
		text := regexp.MustCompile("\\\\[a-zA-Z]+\\d* ?").ReplaceAllString(string(data), " ")
		text = strings.ReplaceAll(text, "{", " ")
		text = strings.ReplaceAll(text, "}", " ")
		extractedText = text
	case ".yaml", ".yml", ".txt", ".json", ".xml":
		extractedText = string(data)
	default:
		return "", fmt.Errorf("unsupported file type: %s", ext)
	}

	return extractedText, nil
}
