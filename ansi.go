package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/glamour"
	"golang.org/x/term"
)

/*
renderMarkdown2Ansi renders markdown to ANSI escape codes for terminal display. It converts a markdown string
to ANSI format suitable for terminal output, adjusting to the terminal width and applying color replacements.
*/
func renderMarkdown2Ansi(md string) string {
	terminalWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		terminalWidth = 80 // fallback width
	} else {
		terminalWidth -= 4 // avoid unwanted line breaks for edge cases
	}
	if terminalWidth < 80 {
		terminalWidth = 80 // fallback width
	}

	// convert markdown data to terminal data
	terminalRenderer, err := glamour.NewTermRenderer(
		glamour.WithStylePath(progConfig.AnsiOutputTheme),
		glamour.WithWordWrap(terminalWidth),
		glamour.WithEmoji(),
	)
	if err != nil {
		message := fmt.Sprintf("error [%v] at glamour.NewTermRenderer()", err)
		fmt.Printf("%s\n", message)
		return message
	}

	defer func() { _ = terminalRenderer.Close() }()

	terminalData, err := terminalRenderer.Render(md)
	if err != nil {
		message := fmt.Sprintf("error [%v] at terminalRenderer.Render(()", err)
		fmt.Printf("%s\n", message)
		return message
	}

	return terminalData
}
