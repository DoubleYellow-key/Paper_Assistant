package parser

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	pdf "github.com/ledongthuc/pdf"
)

func ParseFileText(path string) (string, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".pdf":
		return parsePDF(path)
	case ".txt", ".md":
		return parsePlainText(path)
	default:
		return "", fmt.Errorf("unsupported file type: %s", ext)
	}
}

func parsePDF(path string) (string, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return "", fmt.Errorf("open pdf %s: %w", path, err)
	}
	defer f.Close()

	reader, err := r.GetPlainText()
	if err != nil {
		return "", fmt.Errorf("extract pdf text %s: %w", path, err)
	}

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(reader); err != nil {
		return "", fmt.Errorf("read pdf text %s: %w", path, err)
	}
	return normalizeText(buf.String()), nil
}

func parsePlainText(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read file %s: %w", path, err)
	}
	return normalizeText(string(b)), nil
}

func normalizeText(s string) string {
	s = strings.ReplaceAll(s, "\u0000", "")
	s = strings.TrimSpace(s)
	return s
}
