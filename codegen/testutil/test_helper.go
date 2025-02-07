package testutil

import (
	"strings"
	"testing"
)

// AssemblyMatcher checks if generated assembly contains expected instructions
func AssemblyMatcher(t *testing.T, generated, expected string) {
	t.Helper()
	genLines := cleanAssembly(generated)
	expLines := cleanAssembly(expected)

	for _, expLine := range expLines {
		found := false
		for _, genLine := range genLines {
			if expLine == genLine {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected instruction not found: %s", expLine)
		}
	}
}

// cleanAssembly removes empty lines and normalizes whitespace
func cleanAssembly(asm string) []string {
	var lines []string
	for _, line := range strings.Split(asm, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}
