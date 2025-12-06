package extractor

import (
	"os"
	"testing"
)

func TestExtractPDF(t *testing.T) {
	data, err := os.ReadFile("testdata/sample.pdf")
	if err != nil {
		t.Fatalf("failed to read sample PDF: %v", err)
	}

	text, err := ExtractPDF(data)
	if err != nil {
		t.Fatalf("ExtractPDF returned error: %v", err)
	}

	if text == "" {
		t.Errorf("ExtractPDF returned empty text")
	}

	t.Logf("Extracted PDF text:\n%s", text)
}

func TestExtractDOCX(t *testing.T) {
	data, err := os.ReadFile("testdata/sample.docx")
	if err != nil {
		t.Fatalf("failed to read sample DOCX: %v", err)
	}

	text, err := ExtractDOCX(data)
	if err != nil {
		t.Fatalf("ExtractDOCX returned error: %v", err)
	}

	if text == "" {
		t.Errorf("ExtractDOCX returned empty text")
	}

	t.Logf("Extracted DOCX text:\n%s", text)
}
