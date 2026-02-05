package main

import (
	"regexp"
	"strings"
)

var (
	// markdownBlockRegex detects an outer markdown wrapper.
	// Matches: ```markdown (optional), Newline, Content, Newline, ```.
	markdownBlockRegex = regexp.MustCompile(`(?i)^\x60{3}(?:markdown)?\s*\n(?s)(.*)\n\x60{3}$`)
)

/*
cleanMarkdown normalizes the AI's markdown output for further processing.
It removes technical artifacts (wrappers) and corrects syntax errors,
such as indented code fences, that prevent parsers like Goldmark from rendering correctly.
*/
func cleanMarkdown(input string) string {
	// 1. Remove outer markdown wrapper if present.
	// Many LLMs tend to wrap the entire response in ```markdown ... ```.
	result := unwrapMarkdownBlock(input)

	// 2. Remove whitespace at the beginning and end of the document.
	result = strings.TrimSpace(result)

	// 3. Specific fix for code blocks:
	// Removes spaces before ``` if they immediately follow a newline.
	// Reason: Indented backticks (e.g., 4 spaces) are often interpreted in Markdown
	// as an "Indented Code Block" or quote, breaking syntax highlighting for the
	// actual "Fenced Code Block".
	result = removeSpacesBetweenNewlineAndCodeblock(result)

	return result
}

/*
unwrapMarkdownBlock removes the outer code block frame if the AI
has wrapped the entire response within one.
*/
func unwrapMarkdownBlock(input string) string {
	// Trim to ensure regex matches the start/end correctly.
	trimmed := strings.TrimSpace(input)

	groups := markdownBlockRegex.FindStringSubmatch(trimmed)
	if len(groups) == 2 {
		// Group 1 contains the content inside the backticks.
		return groups[1]
	}
	return input
}

/*
removeSpacesBetweenNewlineAndCodeblock searches for the sequence:
[Newline] + [Spaces] + [```]
and replaces it with:
[Newline] + [```]

This is safer than a global "unindent" because it preserves indentation
within code (e.g., Python) or lists.
*/
func removeSpacesBetweenNewlineAndCodeblock(input string) string {
	var output strings.Builder
	// Optimization: Pre-allocate buffer size.
	output.Grow(len(input))

	length := len(input)
	for i := 0; i < length; i++ {
		// Check for newline.
		if input[i] == '\n' {
			// Lookahead.
			j := i + 1

			// Skip all subsequent spaces.
			for j < length && input[j] == ' ' {
				j++
			}

			// Check if a code block (```) starts after the spaces.
			if j+2 < length && input[j] == '`' && input[j+1] == '`' && input[j+2] == '`' {
				// Match: We found an indented code block.

				// 1. Write the newline.
				output.WriteByte('\n')

				// 2. "Delete" the spaces by setting the main index i to the position
				// BEFORE the backticks (j-1). In the next loop iteration,
				// i will be incremented and land exactly on the first backtick.
				i = j - 1
			} else {
				// No match: It was just an empty line or normal text.
				// Write the newline normally.
				output.WriteByte(input[i])
			}
		} else {
			// Copy normal character.
			output.WriteByte(input[i])
		}
	}
	return output.String()
}
