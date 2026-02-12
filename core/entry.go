package core

import (
	"bytes"
	"fmt"
	"strings"
)

type Entry struct {
	Date    string
	Prompt  string
	Minimum int
	Body    string
}

// ParseEntry parses a markdown file with YAML frontmatter into an Entry.
func ParseEntry(raw []byte) (Entry, error) {
	content := string(raw)
	var e Entry

	if !strings.HasPrefix(content, "---\n") {
		return Entry{}, fmt.Errorf("missing frontmatter")
	}

	end := strings.Index(content[4:], "\n---\n")
	if end == -1 {
		return Entry{}, fmt.Errorf("unterminated frontmatter")
	}

	frontmatter := content[4 : 4+end]
	e.Body = strings.TrimPrefix(content[4+end+5:], "\n")

	for _, line := range strings.Split(frontmatter, "\n") {
		key, val, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		val = strings.Trim(val, "\"")

		switch key {
		case "date":
			e.Date = val
		case "prompt":
			e.Prompt = val
		case "minimum":
			n := 0
			fmt.Sscanf(val, "%d", &n)
			e.Minimum = n
		}
	}

	return e, nil
}

// SerializeEntry writes an Entry back to markdown with YAML frontmatter.
func SerializeEntry(e Entry) []byte {
	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.WriteString(fmt.Sprintf("date: %s\n", e.Date))
	if e.Prompt != "" {
		buf.WriteString(fmt.Sprintf("prompt: \"%s\"\n", e.Prompt))
	}
	buf.WriteString(fmt.Sprintf("minimum: %d\n", e.Minimum))
	buf.WriteString("---\n\n")
	buf.WriteString(e.Body)
	return buf.Bytes()
}

// WordCount counts words by splitting on whitespace.
func WordCount(text string) int {
	fields := strings.Fields(text)
	return len(fields)
}
