package scanner

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// ExtractText converts various document formats into plain text.
func ExtractText(data []byte, filename string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	var extractedText string
	switch ext {
	case ".pdf":
		var sb strings.Builder
		re := regexp.MustCompile(`\(([^)]*)\)\s*T[Jj]`)
		for _, m := range re.FindAllSubmatch(data, -1) {
			sb.Write(m[1])
		}
		extractedText = sb.String()
	case ".html", ".htm":
		re := regexp.MustCompile("<[^>]+>")
		extractedText = re.ReplaceAllString(string(data), " ")
		case ".yaml", ".yml", ".txt", ".json", ".xml":
		extractedText = string(data)
	default:
		return "", fmt.Errorf("unsupported file type: %s", ext)
	}

	return extractedText, nil
}
