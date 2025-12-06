package extractor

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

type WordDocument struct {
	XMLName xml.Name `xml:"document"`
	Body    Body     `xml:"body"`
}

type Body struct {
	Paragraphs []Paragraph `xml:"p"`
}

type Paragraph struct {
	Runs []Run `xml:"r"`
}

type Run struct {
	Text string `xml:"t"`
}

func ExtractDOCX(data []byte) (string, error) {
	reader := bytes.NewReader(data)

	zipReader, err := zip.NewReader(reader, int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("failed to read DOCX as ZIP: %w", err)
	}

	// Find document.xml
	var documentFile *zip.File
	for _, file := range zipReader.File {
		if file.Name == "word/document.xml" {
			documentFile = file
			break
		}
	}

	if documentFile == nil {
		return "", fmt.Errorf("document.xml not found in DOCX")
	}

	// Read document.xml
	xmlFile, err := documentFile.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open document.xml: %w", err)
	}
	defer xmlFile.Close()

	xmlData, err := io.ReadAll(xmlFile)
	if err != nil {
		return "", fmt.Errorf("failed to read document.xml: %w", err)
	}

	// Parse XML
	var doc WordDocument
	if err := xml.Unmarshal(xmlData, &doc); err != nil {
		return "", fmt.Errorf("failed to parse document.xml: %w", err)
	}

	// Extract text
	var textBuilder strings.Builder
	for _, para := range doc.Body.Paragraphs {
		for _, run := range para.Runs {
			textBuilder.WriteString(run.Text)
		}
		textBuilder.WriteString("\n")
	}

	extractedText := strings.TrimSpace(textBuilder.String())

	if extractedText == "" {
		return "", fmt.Errorf("no text could be extracted from DOCX")
	}

	return extractedText, nil
}
