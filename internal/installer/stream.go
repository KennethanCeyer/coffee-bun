package installer

import (
	"bytes"
	"regexp"
	"strings"
	"sync"
)

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
var emojiRegex = regexp.MustCompile(`[\x{1F600}-\x{1F64F}\x{1F300}-\x{1F5FF}\x{1F680}-\x{1F6FF}\x{1F700}-\x{1F77F}\x{1F780}-\x{1F7FF}\x{1F800}-\x{1F8FF}\x{1F900}-\x{1F9FF}\x{1FA00}-\x{1FA6F}\x{1FA70}-\x{1FAFF}\x{2600}-\x{26FF}\x{2700}-\x{27BF}\x{2300}-\x{23FF}\x{2B50}]`)

func cleanLogLine(s string) string {
	s = ansiRegex.ReplaceAllString(s, "")
	s = emojiRegex.ReplaceAllString(s, "")
	s = strings.TrimSpace(s)
	if len(s) > 90 {
		s = s[:87] + "..."
	}
	return s
}

// lineWriter captures stdout/stderr and streams line updates in real time
type lineWriter struct {
	mu       sync.Mutex
	buf      bytes.Buffer
	rem      string
	lastLine string
	onLine   func(string)
}

func newLineWriter(onLine func(string)) *lineWriter {
	return &lineWriter{
		onLine: onLine,
	}
}

func (lw *lineWriter) Write(p []byte) (n int, err error) {
	lw.mu.Lock()
	defer lw.mu.Unlock()

	lw.buf.Write(p)
	str := lw.rem + string(p)
	str = strings.ReplaceAll(str, "\r\n", "\n")

	parts := strings.FieldsFunc(str, func(r rune) bool {
		return r == '\n' || r == '\r'
	})

	if len(parts) == 0 {
		return len(p), nil
	}

	hasEndDelimiter := strings.HasSuffix(str, "\n") || strings.HasSuffix(str, "\r")

	var completeParts []string
	if hasEndDelimiter {
		completeParts = parts
		lw.rem = ""
	} else {
		completeParts = parts[:len(parts)-1]
		lw.rem = parts[len(parts)-1]
	}

	for _, rawLine := range completeParts {
		cleaned := cleanLogLine(rawLine)
		if cleaned != "" {
			lw.lastLine = cleaned
			if lw.onLine != nil {
				lw.onLine(cleaned)
			}
		}
	}

	return len(p), nil
}

func (lw *lineWriter) Flush() {
	lw.mu.Lock()
	defer lw.mu.Unlock()

	if lw.rem != "" {
		cleaned := cleanLogLine(lw.rem)
		if cleaned != "" {
			lw.lastLine = cleaned
			if lw.onLine != nil {
				lw.onLine(cleaned)
			}
		}
		lw.rem = ""
	}
}

func (lw *lineWriter) LastLine() string {
	lw.mu.Lock()
	defer lw.mu.Unlock()
	return lw.lastLine
}
