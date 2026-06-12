package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/aquilax/truncate"
)

/*
parseMarkdownToHTML is a helper function to render a piece of markdown to html.
*/
func parseMarkdownToHTML(md string) string {
	var buf bytes.Buffer
	err := markdownParser.Convert([]byte(md), &buf)
	if err != nil {
		fmt.Printf("error [%v] at markdownParser.Convert()", err)
	}
	return buf.String()
}

/*
renderMarkdown2HTML renders markdown text to HTML format. It converts a markdown string into HTML by
parsing it and applying configured HTML element replacements for output formatting.
It uses a "Divide and Conquer" strategy to separate logical units before parsing to prevent syntax
errors in one section from breaking the overall HTML structure.
*/
func renderMarkdown2HTML(md string) string {
	type blockDef struct {
		start string
		end   string
		title string
		open  bool
	}

	defs := []blockDef{
		{"<!-- THOUGHTS_START -->", "<!-- THOUGHTS_END -->", "Thoughts", false},
		{"<!-- PROMPT_USER_START -->", "<!-- PROMPT_USER_END -->", "User Prompt", false},
		{"<!-- PROMPT_SYSTEM_START -->", "<!-- PROMPT_SYSTEM_END -->", "System Prompt", false},
		{"<!-- PROMPT_RESOURCES_START -->", "<!-- PROMPT_RESOURCES_END -->", "Resources", false},
		{"<!-- RESPONSE_START -->", "<!-- RESPONSE_END -->", "Response", true},
		{"<!-- STATS_START -->", "<!-- STATS_END -->", "Statistics", false},
	}

	var htmlBuilder strings.Builder
	remaining := md

	for {
		firstMarkerIdx := -1
		var activeDef blockDef

		for _, def := range defs {
			idx := strings.Index(remaining, def.start)
			if idx != -1 {
				if firstMarkerIdx == -1 || idx < firstMarkerIdx {
					firstMarkerIdx = idx
					activeDef = def
				}
			}
		}

		if firstMarkerIdx == -1 {
			// parse remaining text
			htmlBuilder.WriteString(parseMarkdownToHTML(remaining))
			break
		}

		// parse text before the marker
		before := remaining[:firstMarkerIdx]
		htmlBuilder.WriteString(parseMarkdownToHTML(before))

		// find the end marker
		endSearchArea := remaining[firstMarkerIdx+len(activeDef.start):]
		relativeEndIdx := strings.Index(endSearchArea, activeDef.end)

		// determine if the details element should be open by default
		openAttr := ""
		if activeDef.open {
			openAttr = " open"
		}

		if relativeEndIdx == -1 {
			// missing end marker, parse the rest as inner block
			inner := endSearchArea
			fmt.Fprintf(&htmlBuilder, "<details%s><summary>%s</summary>\n%s\n</details>\n", openAttr, activeDef.title, parseMarkdownToHTML(inner))
			break
		}

		// parse inner block
		inner := endSearchArea[:relativeEndIdx]
		fmt.Fprintf(&htmlBuilder, "<details%s><summary>%s</summary>\n%s\n</details>\n", openAttr, activeDef.title, parseMarkdownToHTML(inner))

		// advance remaining string
		remaining = endSearchArea[relativeEndIdx+len(activeDef.end):]
	}

	htmlDataModified := htmlBuilder.String()

	// replace HTML elements
	for _, item := range progConfig.HTMLReplaceElements {
		for key, value := range item {
			htmlDataModified = strings.ReplaceAll(htmlDataModified, key, value)
		}
	}

	return htmlDataModified
}

/*
buildHTMLPage constructs a complete HTML page by combining a header, body, and footer. It reads an HTML body
from a source file, combines it with header and footer content from configuration, and writes the complete
HTML page to a destination file.
*/
func buildHTMLPage(title, source, destination string) error {
	htmlBody, err := os.ReadFile(source)
	if err != nil {
		fmt.Printf("error [%v] at os.ReadFile()", err)
		return err
	}

	title = strings.ReplaceAll(title, "\r\n", " ")
	title = strings.ReplaceAll(title, "\n", " ")
	title = strings.ReplaceAll(title, "\t", " ")

	title = truncate.Truncate(title, progConfig.HTMLMaxLengthTitle, "...", truncate.PositionEnd)
	htmlHeader := fmt.Sprintf(progConfig.HTMLHeader, title)
	htmlFooter := progConfig.HTMLFooter

	// build html page
	htmlPage := htmlHeader + string(htmlBody) + htmlFooter

	// write html to file
	err = os.WriteFile(destination, []byte(htmlPage), 0600)
	if err != nil {
		fmt.Printf("error [%v] at os.WriteFile()", err)
		return err
	}

	return nil
}
