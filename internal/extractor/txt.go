package extractor

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

func ExtractTXT(data []byte) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("empty text file")
	}

	text, err := decodeText(data)
	if err != nil {
		return "", fmt.Errorf("failed to decode text file: %w", err)
	}

	text = cleanText(text)

	if text == "" {
		return "", fmt.Errorf("no text could be extracted from file")
	}

	return text, nil
}

func decodeText(data []byte) (string, error) {
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		return string(data[3:]), nil
	}

	if len(data) >= 2 && data[0] == 0xFF && data[1] == 0xFE {
		decoder := unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder()
		decoded, _, err := transform.Bytes(decoder, data)
		if err != nil {
			return "", err
		}
		return string(decoded), nil
	}

	if len(data) >= 2 && data[0] == 0xFE && data[1] == 0xFF {
		decoder := unicode.UTF16(unicode.BigEndian, unicode.UseBOM).NewDecoder()
		decoded, _, err := transform.Bytes(decoder, data)
		if err != nil {
			return "", err
		}
		return string(decoded), nil
	}

	if utf8.Valid(data) {
		return string(data), nil
	}

	decoder := charmap.Windows1252.NewDecoder()
	decoded, _, err := transform.Bytes(decoder, data)
	if err == nil {
		return string(decoded), nil
	}

	decoder = charmap.ISO8859_1.NewDecoder()
	decoded, _, err = transform.Bytes(decoder, data)
	if err == nil {
		return string(decoded), nil
	}

	return string(data), nil
}

func cleanText(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	text = strings.ReplaceAll(text, "\x00", "")

	lines := strings.Split(text, "\n")

	var cleanedLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			cleanedLines = append(cleanedLines, line)
		}
	}

	result := strings.Join(cleanedLines, "\n")

	return strings.TrimSpace(result)
}

func ExtractTXTSimple(data []byte) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("empty text file")
	}

	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		data = data[3:]
	}

	text := string(data)

	text = cleanText(text)

	if text == "" {
		return "", fmt.Errorf("no text could be extracted from file")
	}

	return text, nil
}

// ValidateTXT checks if the data appears to be valid text
func ValidateTXT(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("empty file")
	}

	// Check if it's mostly printable or whitespace characters
	printableCount := 0
	sampleSize := 512
	if len(data) < sampleSize {
		sampleSize = len(data)
	}

	for i := 0; i < sampleSize; i++ {
		b := data[i]
		// Printable ASCII, tabs, newlines, carriage returns
		if (b >= 32 && b <= 126) || b == '\t' || b == '\n' || b == '\r' {
			printableCount++
		}
	}

	// If less than 80% of sample is printable text, it might be binary
	if float64(printableCount)/float64(sampleSize) < 0.8 {
		return fmt.Errorf("file does not appear to be valid text")
	}

	return nil
}
