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
	switch ext {
	case ".pdf":
		var sb strings.Builder
		re := regexp.MustCompile(`\(([^)]*)\)\s*T[Jj]`)
		for _, m := range re.FindAllSubmatch(data, -1) {
			sb.Write(m[1])
		}
		return sb.String(), nil
	case ".html", ".htm":
		re := regexp.MustCompile("<[^>]+>")
		return re.ReplaceAllString(string(data), " "), nil
	case ".yaml", ".yml", ".txt":
		return string(data), nil
	default:
		return "", fmt.Errorf("unsupported file type: %s", ext)
	}
}
