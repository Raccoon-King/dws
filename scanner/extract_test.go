package scanner

import (
	"strings"
	"testing"
)

func TestExtractTextPDF(t *testing.T) {
	// Skip this test if we can't create a valid PDF for testing
	// The actual PDF parsing functionality is tested with real PDFs in integration tests
	t.Skip("PDF extraction requires a valid PDF file - tested in integration")
}

func TestExtractTextHTML(t *testing.T) {
	data := []byte("<html><body><p>hi</p></body></html>")
	txt, err := ExtractText(data, "file.html")
	if err != nil || strings.TrimSpace(txt) != "hi" {
		t.Fatalf("unexpected: %v %q", err, txt)
	}
}

func TestExtractTextYAML(t *testing.T) {
	data := []byte("a: 1")
	txt, err := ExtractText(data, "file.yaml")
	if err != nil || txt != "a: 1" {
		t.Fatalf("unexpected: %v %q", err, txt)
	}
}

func TestExtractTextJSON(t *testing.T) {
	data := []byte(`{"key": "value"}`)
	txt, err := ExtractText(data, "file.json")
	if err != nil || txt != `{"key": "value"}` {
		t.Fatalf("unexpected: %v %q", err, txt)
	}
}

func TestExtractTextXML(t *testing.T) {
	data := []byte("<root><item>data</item></root>")
	txt, err := ExtractText(data, "file.xml")
	if err != nil || txt != "<root><item>data</item></root>" {
		t.Fatalf("unexpected: %v %q", err, txt)
	}
}

func TestExtractTextHTM(t *testing.T) {
	data := []byte("<html><body><p>HTM test</p></body></html>")
	txt, err := ExtractText(data, "file.htm")
	if err != nil || strings.TrimSpace(txt) != "HTM test" {
		t.Fatalf("unexpected: %v %q", err, txt)
	}
}

func TestExtractTextYML(t *testing.T) {
	data := []byte("key: value\narray:\n  - item1\n  - item2")
	txt, err := ExtractText(data, "file.yml")
	if err != nil || txt != "key: value\narray:\n  - item1\n  - item2" {
		t.Fatalf("unexpected: %v %q", err, txt)
	}
}

func TestExtractTextTXT(t *testing.T) {
	data := []byte("This is plain text.")
	txt, err := ExtractText(data, "file.txt")
	if err != nil || txt != "This is plain text." {
		t.Fatalf("unexpected: %v %q", err, txt)
	}
}

func TestExtractTextUnsupported(t *testing.T) {
	// Test with data that contains null bytes (binary)
	if _, err := ExtractText([]byte("\x00\x01\x02\x03"), "file.bin"); err == nil {
		t.Fatalf("expected error for binary data")
	}
}

func TestExtractTextBMP(t *testing.T) {
	// Test BMP file (binary format)
	if _, err := ExtractText([]byte("\x42\x4D\x94\x87\x00\x00"), "file.bmp"); err == nil {
		t.Fatalf("expected error for BMP file")
	}
}

func TestExtractTextUnknown(t *testing.T) {
	// Test unknown extension but text content
	data := []byte("This is text content in unknown format")
	txt, err := ExtractText(data, "file.unknown")
	if err != nil || txt != "This is text content in unknown format" {
		t.Fatalf("unexpected: %v %q", err, txt)
	}
}

func TestExtractTextNoExtension(t *testing.T) {
	// Test file with no extension
	data := []byte("No extension text")
	txt, err := ExtractText(data, "file-no-ext")
	if err != nil || txt != "No extension text" {
		t.Fatalf("unexpected: %v %q", err, txt)
	}
}

func TestExtractTextMixedContent(t *testing.T) {
	data := []byte("normal text\x00with null")
	txt, err := ExtractText(data, "file.txt")
	if err != nil || txt != "normal text\x00with null" {
		t.Fatalf("unexpected: %v %q", err, txt)
	}
}

func TestHTMLTagRemoval(t *testing.T) {
	html := `<html><head><title>Test</title></head><body><h1>Hello</h1><p>This is a paragraph</p></body></html>`
	txt, err := ExtractText([]byte(html), "file.html")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	txt = strings.TrimSpace(txt)
	if !strings.Contains(txt, "Test") || !strings.Contains(txt, "Hello") || !strings.Contains(txt, "paragraph") {
		t.Fatalf("HTML tag removal failed: expected content not found in: %q", txt)
	}
}

func TestExtractTextEmptyFile(t *testing.T) {
	data := []byte("")
	txt, err := ExtractText(data, "empty.txt")
	if err != nil || txt != "" {
		t.Fatalf("unexpected for empty file: %v %q", err, txt)
	}
}

func TestBinaryDataDetection(t *testing.T) {
	testCases := []struct {
		name     string
		data     []byte
		expected bool
	}{
		{"empty", []byte(""), false},
		{"text only", []byte("hello world"), false},
		{"with null byte", []byte("hello\x00world"), true},
		{"leading null", []byte("\x00hello"), true},
		{"binary header", []byte("\x89\x50\x4E\x47"), true},
		{"text with null", []byte("test\x00123"), true},
	}

	for _, tc := range testCases {
		result := isBinaryData(tc.data)
		if result != tc.expected {
			t.Errorf("%s: expected %v, got %v", tc.name, tc.expected, result)
		}
	}
}
