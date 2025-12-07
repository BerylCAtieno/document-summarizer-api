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
	Content []BodyElement
}

type BodyElement struct {
	Paragraph *Paragraph `xml:"p"`
	Table     *Table     `xml:"tbl"`
}

type Paragraph struct {
	Runs []Run `xml:"r"`
}

type Table struct {
	Rows []TableRow `xml:"tr"`
}

type TableRow struct {
	Cells []TableCell `xml:"tc"`
}

type TableCell struct {
	Paragraphs []Paragraph `xml:"p"`
}

type Run struct {
	Text     string   `xml:"t"`
	TabChar  *TabChar `xml:"tab"`
	BreakTag *Break   `xml:"br"`
}

type TabChar struct{}

type Break struct{}

// UnmarshalXML implements custom unmarshaling for Body to capture mixed content
func (b *Body) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	for {
		token, err := d.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		switch elem := token.(type) {
		case xml.StartElement:
			var be BodyElement
			switch elem.Name.Local {
			case "p":
				var p Paragraph
				if err := d.DecodeElement(&p, &elem); err != nil {
					return err
				}
				be.Paragraph = &p
				b.Content = append(b.Content, be)
			case "tbl":
				var t Table
				if err := d.DecodeElement(&t, &elem); err != nil {
					return err
				}
				be.Table = &t
				b.Content = append(b.Content, be)
			case "sectPr":
				// Skip section properties
				d.Skip()
			default:
				// Skip unknown elements
				d.Skip()
			}
		case xml.EndElement:
			if elem.Name.Local == "body" {
				return nil
			}
		}
	}
	return nil
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

	for _, element := range doc.Body.Content {
		if element.Paragraph != nil {
			extractParagraph(element.Paragraph, &textBuilder)
			textBuilder.WriteString("\n")
		} else if element.Table != nil {
			extractTable(element.Table, &textBuilder)
			textBuilder.WriteString("\n")
		}
	}

	extractedText := strings.TrimSpace(textBuilder.String())

	if extractedText == "" {
		return "", fmt.Errorf("no text could be extracted from DOCX")
	}

	return extractedText, nil
}

func extractParagraph(para *Paragraph, builder *strings.Builder) {
	for _, run := range para.Runs {
		if run.Text != "" {
			builder.WriteString(run.Text)
		}
		if run.TabChar != nil {
			builder.WriteString("\t")
		}
		if run.BreakTag != nil {
			builder.WriteString("\n")
		}
	}
}

func extractTable(table *Table, builder *strings.Builder) {
	for _, row := range table.Rows {
		var cellTexts []string
		for _, cell := range row.Cells {
			var cellBuilder strings.Builder
			for _, para := range cell.Paragraphs {
				extractParagraph(&para, &cellBuilder)
			}
			cellText := strings.TrimSpace(cellBuilder.String())
			if cellText != "" {
				cellTexts = append(cellTexts, cellText)
			}
		}
		if len(cellTexts) > 0 {
			builder.WriteString(strings.Join(cellTexts, " | "))
			builder.WriteString("\n")
		}
	}
}
