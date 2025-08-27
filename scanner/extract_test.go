package scanner

import (
	"strings"
	"testing"
)

func TestExtractTextPDF(t *testing.T) {
	data := []byte("%PDF-1.1\nBT\n(hello) Tj\nET\n%%EOF")
	txt, err := ExtractText(data, "test.pdf")
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	if strings.TrimSpace(txt) != "hello" {
		t.Fatalf("expected text, got %q", txt)
	}
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

func TestExtractTextUnsupported(t *testing.T) {
	if _, err := ExtractText([]byte(""), "file.bin"); err == nil {
		t.Fatalf("expected error")
	}
}
