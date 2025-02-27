package preprocessor

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Preprocessor struct {
	processedFiles map[string]bool
	originalFile   string // Track the original file being compiled
}

func New() *Preprocessor {
	return &Preprocessor{
		processedFiles: make(map[string]bool),
		originalFile:   "",
	}
}

func (p *Preprocessor) ProcessFile(filename string) (string, error) {
	if p.originalFile == "" {
		p.originalFile = filename
	}

	if p.processedFiles[filename] {
		return "", fmt.Errorf("circular import detected: %s", filename)
	}

	p.processedFiles[filename] = true

	content, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}

	processed := ""
	scanner := bufio.NewScanner(strings.NewReader(string(content)))

	inMainClass := false
	mainClassDepth := 0

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		if strings.HasPrefix(trimmedLine, "import") {
			// Extract module name from import statement
			parts := strings.Split(strings.TrimSpace(line), " ")
			if len(parts) != 2 {
				return "", fmt.Errorf("invalid import statement: %s", line)
			}

			moduleName := strings.TrimSuffix(parts[1], ";")
			moduleFile := filepath.Join(filepath.Dir(filename), moduleName+".cl")

			// Process the imported file
			importedContent, err := p.ProcessFile(moduleFile)
			if err != nil {
				return "", fmt.Errorf("error processing import %s: %v", moduleFile, err)
			}

			processed += importedContent + "\n"
		} else {
			// Check for main class start
			if strings.HasPrefix(trimmedLine, "class Main") {
				inMainClass = true
				mainClassDepth = 0
				// Skip this line if we're processing an imported file
				if filename != p.originalFile {
					continue
				}
			}

			// Count braces to track class scope
			if inMainClass {
				mainClassDepth += strings.Count(line, "{")
				mainClassDepth -= strings.Count(line, "}")

				// If we've reached the end of main class
				if mainClassDepth < 0 {
					inMainClass = false
					// Skip this line if we're processing an imported file
					if filename != p.originalFile {
						continue
					}
				}

				// Skip all lines within main class for imported files
				if filename != p.originalFile {
					continue
				}
			}

			processed += line + "\n"
		}
	}

	return processed, nil
}
