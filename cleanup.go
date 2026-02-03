/*
Markdown Cleaning Utilities
This file provides functions to sanitize and normalize Markdown output received from
AI models. It handles common issues like unnecessary outer wrappers, inconsistent
indentation in code blocks, and the preservation of structural Markdown elements.
*/

package main

import (
	"bufio"
	"regexp"
	"strings"
)

var (
	listItemRegex      = regexp.MustCompile(`^(\d+[\.\)]|\*|-|\+)\s`)
	definitionRegex    = regexp.MustCompile(`^:\s`)
	markdownBlockRegex = regexp.MustCompile(`(?i)^\x60{3}(?:markdown)?\s*\n(?s)(.*)\n\x60{3}$`)
)

/*
cleanMarkdownIndentation normalizes the indentation of a Markdown string.
It removes accidental leading whitespace from standard text while carefully
preserving structural indentation required for lists, blockquotes, and tables.

It implements "Smart Dedent" for fenced code blocks: if a code block is indented,
the content within is shifted left relative to the opening backticks.

It also detects and preserves indentation within LaTeX/Math blocks ($$ ... $$
and \[ ... \]) to ensure mathematical formulas retain their formatting.

Example:

	Input:
	  "    This is a paragraph."
	  "    * List item"
	  "    ```go"
	  "        fmt.Println(1)"
	  "    ```"

	Output:
	  "This is a paragraph."
	  "    * List item" (Preserved)
	  "```go"
	  "    fmt.Println(1)" (Shifted left by 4 spaces)
	  "```"
*/
func cleanMarkdownIndentation(markdown string) string {
	var result strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(markdown))

	inFencedCodeBlock := false
	inMathBlock := false
	currentBlockIndent := 0

	for scanner.Scan() {
		originalLine := scanner.Text()

		// Preserve Hard Line Breaks (standard Markdown: two spaces at end of line)
		trimmedLine := strings.TrimRight(originalLine, " \t")
		if strings.HasSuffix(originalLine, "  ") && len(trimmedLine) > 0 {
			trimmedLine += "  "
		}

		if len(strings.TrimSpace(trimmedLine)) == 0 {
			result.WriteString("\n")
			continue
		}

		contentCheck := strings.TrimSpace(trimmedLine)

		// Handle Fenced Code Blocks
		if strings.HasPrefix(contentCheck, "```") {
			if !inFencedCodeBlock {
				inFencedCodeBlock = true
				currentBlockIndent = strings.Index(originalLine, "`")
				if currentBlockIndent < 0 {
					currentBlockIndent = 0
				}
			} else {
				inFencedCodeBlock = false
				currentBlockIndent = 0
			}
			result.WriteString(contentCheck + "\n")
			continue
		}

		// Handle LaTeX/Math Blocks
		// Case 1: $$ Block (Display Math)
		if strings.HasPrefix(contentCheck, "$$") {
			// Check: Is it a single-line block? (e.g. "$$ E=mc^2 $$")
			if len(contentCheck) > 2 && strings.HasSuffix(contentCheck, "$$") {
				// Single-liner -> Keep status, just write.
				// We use trimmedLine to preserve potential leading whitespace (indentation).
				result.WriteString(trimmedLine + "\n")
				continue
			}
			// It is a multi-liner toggle (start or end)
			inMathBlock = !inMathBlock
			result.WriteString(trimmedLine + "\n")
			continue
		}
		// Case 2: \[ ... \] Block (Display Math)
		if strings.HasPrefix(contentCheck, "\\[") {
			// Check: Is it a single-line block? (e.g. "\[ a^2 + b^2 = c^2 \]")
			if strings.HasSuffix(contentCheck, "\\]") {
				// Single-liner -> Do not change status
				result.WriteString(trimmedLine + "\n")
				continue
			}
			// Start of a multi-line block
			inMathBlock = true
			result.WriteString(trimmedLine + "\n")
			continue
		}
		// Case 3: \] End of a block
		if strings.HasPrefix(contentCheck, "\\]") {
			inMathBlock = false
			result.WriteString(trimmedLine + "\n")
			continue
		}

		switch {
		case inFencedCodeBlock:
			// Apply Smart Dedent: Remove the same amount of space as the opening backticks
			var linePayload string
			prefixSpaceCount := 0
			for _, r := range originalLine {
				if r == ' ' {
					prefixSpaceCount++
				} else {
					break
				}
			}

			if prefixSpaceCount >= currentBlockIndent {
				linePayload = originalLine[currentBlockIndent:]
			} else {
				linePayload = strings.TrimLeft(originalLine, " ")
			}
			result.WriteString(strings.TrimRight(linePayload, " \t") + "\n")

		case inMathBlock:
			result.WriteString(trimmedLine + "\n")

		default:
			// Handle standard text
			leftTrimmed := strings.TrimLeft(trimmedLine, " ")

			if shouldPreserveIndentation(leftTrimmed) {
				result.WriteString(trimmedLine + "\n")
			} else {
				result.WriteString(leftTrimmed + "\n")
			}
		}
	}
	return strings.TrimSuffix(result.String(), "\n")
}

/*
shouldPreserveIndentation determines if a line's leading spaces are semantically
meaningful in Markdown. It returns true for:
- Blockquotes (starting with >)
- Tables (starting with |)
- Footnotes (starting with [^)
- Horizontal rules (---, ***)
- Ordered and unordered lists
- HTML blocks
- Definition lists
*/
func shouldPreserveIndentation(line string) bool {
	if strings.HasPrefix(line, ">") ||
		strings.HasPrefix(line, "|") ||
		strings.HasPrefix(line, "[^") ||
		strings.HasPrefix(line, "<") {
		return true
	}
	if isHorizontalRule(line) ||
		listItemRegex.MatchString(line) ||
		definitionRegex.MatchString(line) {
		return true
	}
	return false
}

/*
isHorizontalRule checks if a line represents a Markdown horizontal rule.
Example: ---, ***, ___
*/
func isHorizontalRule(line string) bool {
	if strings.HasPrefix(line, "---") ||
		strings.HasPrefix(line, "***") ||
		strings.HasPrefix(line, "___") {
		return true
	}
	return false
}

/*
unwrapMarkdownBlock removes "outer" markdown code wrappers that LLMs
frequently use to encapsulate their entire response.

Example:

	Input:  "```markdown\n# Title\nContent\n```"
	Output: "# Title\nContent"
*/
func unwrapMarkdownBlock(input string) string {
	trimmed := strings.TrimSpace(input)
	groups := markdownBlockRegex.FindStringSubmatch(trimmed)
	if len(groups) == 2 {
		return groups[1]
	}
	return input
}

/*
removeSpacesBetweenNewlineAndCodeblock eliminates leading whitespace specifically
before backtick markers that follow a newline. This fixes rendering issues in many
Markdown parsers where an indented backtick block is treated as a literal paragraph
instead of a fenced code block.

Example:

	Input:  "\n    ```bash"
	Output: "\n```bash"
*/
func removeSpacesBetweenNewlineAndCodeblock(input string) string {
	var output strings.Builder
	length := len(input)
	for i := 0; i < length; i++ {
		if input[i] == '\n' {
			j := i + 1
			for j < length && input[j] == ' ' {
				j++
			}

			if j+2 < length && input[j:j+3] == "```" {
				output.WriteByte('\n')
				i = j - 1
			} else {
				output.WriteByte(input[i])
			}
		} else {
			output.WriteByte(input[i])
		}
	}
	return output.String()
}
