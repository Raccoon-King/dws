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
	if _, err := ExtractText([]byte(""), "file.bin"); err == nil {
		t.Fatalf("expected error")
	}
}
