package main

import (
	"fmt"

	"github.com/charmbracelet/glamour"
)

/*
renderMarkdown2Ansi renders markdown to ANSI escape codes for terminal display. It converts a markdown string
to ANSI format suitable for terminal output, adjusting to the terminal width and applying color replacements.
*/
func renderMarkdown2Ansi(md string) string {
	terminalRenderer, err := glamour.NewTermRenderer(
		glamour.WithStylePath(progConfig.AnsiOutputTheme),
		glamour.WithWordWrap(progConfig.AnsiOutputLineLength),
		glamour.WithEmoji(),
	)
	if err != nil {
		return fmt.Sprintf("error initializing renderer: %v", err)
	}
	defer func() { _ = terminalRenderer.Close() }()

	terminalData, err := terminalRenderer.Render(md)
	if err != nil {
		return fmt.Sprintf("error rendering markdown: %v", err)
	}

	return terminalData
}
