package main

import (
	"os"
	"strings"

	markdown "github.com/Klaus-Tockloth/go-term-markdown"
	"golang.org/x/term"
)

/*
renderMarkdown2Ansi renders markdown to ANSI escape codes for terminal display. It converts a markdown string
to ANSI format suitable for terminal output, adjusting to the terminal width and applying color replacements.
*/
func renderMarkdown2Ansi(md string) string {
	// convert markdown data to terminal data
	terminalWidth, _, _ := term.GetSize(int(os.Stdout.Fd()))
	terminalData := markdown.Render(md, terminalWidth, 0)

	// replace ANSI colors in terminal data
	terminalDataModified := string(terminalData)

	for _, item := range progConfig.AnsiReplaceColors {
		for key, value := range item {
			terminalDataModified = strings.ReplaceAll(terminalDataModified, key, value)
		}
	}

	return terminalDataModified
}
