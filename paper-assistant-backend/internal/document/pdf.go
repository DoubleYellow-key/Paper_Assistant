package document

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/ledongthuc/pdf"
)

func ExtractPDFText(path string) (string, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return "", fmt.Errorf("open pdf: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()

	reader, err := r.GetPlainText()
	if err != nil {
		return "", fmt.Errorf("extract plain text: %w", err)
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, reader); err != nil {
		return "", fmt.Errorf("read extracted pdf text: %w", err)
	}

	return normalizeText(buf.String()), nil
}

func normalizeText(raw string) string {
	raw = strings.ReplaceAll(raw, "\r\n", "\n")
	raw = strings.ReplaceAll(raw, "\r", "\n")
	lines := strings.Split(raw, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}

	var out []string
	blankCount := 0
	for _, line := range lines {
		if line == "" {
			blankCount++
			if blankCount > 1 {
				continue
			}
			out = append(out, "")
			continue
		}
		blankCount = 0
		out = append(out, line)
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}
