package document

import "strings"

func SplitTextIntoChunks(text string, maxChunkSize int) []string {
	paragraphs := strings.Split(text, "\n\n")
	chunks := make([]string, 0)
	var current strings.Builder

	flush := func() {
		if current.Len() == 0 {
			return
		}
		chunks = append(chunks, strings.TrimSpace(current.String()))
		current.Reset()
	}

	for _, paragraph := range paragraphs {
		paragraph = strings.TrimSpace(paragraph)
		if paragraph == "" {
			continue
		}
		if len([]rune(paragraph)) > maxChunkSize {
			flush()
			runes := []rune(paragraph)
			for start := 0; start < len(runes); start += maxChunkSize {
				end := start + maxChunkSize
				if end > len(runes) {
					end = len(runes)
				}
				chunks = append(chunks, strings.TrimSpace(string(runes[start:end])))
			}
			continue
		}
		candidate := paragraph
		if current.Len() > 0 {
			candidate = current.String() + "\n\n" + paragraph
		}
		if len([]rune(candidate)) > maxChunkSize {
			flush()
		}
		if current.Len() > 0 {
			current.WriteString("\n\n")
		}
		current.WriteString(paragraph)
	}
	flush()
	return chunks
}
