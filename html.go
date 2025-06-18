package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/aquilax/truncate"
)

/*
renderMarkdown2HTML renders markdown text to HTML format. It converts a markdown string into HTML by
parsing it and applying configured HTML element replacements for output formatting.
*/
func renderMarkdown2HTML(md string) string {
	// convert markdown data to html
	var buf bytes.Buffer
	err := markdownParser.Convert([]byte(md), &buf)
	if err != nil {
		fmt.Printf("error [%v] at markdownParser.Convert()", err)
	}

	htmlDataModified := string(buf.String())

	// wrap 'Gemini AI thoughts' in HTML object '<details><summary> ... </summary> ...</details>'
	htmlDataModified = strings.ReplaceAll(htmlDataModified, "<!-- AI_THOUGHT_SUMMARY_START -->", "<details><summary>Thoughts - Considerations for answering the prompt:")
	htmlDataModified = strings.ReplaceAll(htmlDataModified, "<!-- AI_THOUGHT_SUMMARY_END -->", "</summary>")
	htmlDataModified = strings.ReplaceAll(htmlDataModified, "<!-- AI_THOUGHT_CONTENT_START -->", "") // remove marker
	htmlDataModified = strings.ReplaceAll(htmlDataModified, "<!-- AI_THOUGHT_CONTENT_END -->", "</details>")

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
func buildHTMLPage(prompt, source, destination string) error {
	htmlBody, err := os.ReadFile(source)
	if err != nil {
		fmt.Printf("error [%v] at os.ReadFile()", err)
		return err
	}

	title := strings.ReplaceAll(prompt, "\r\n", " ")
	title = strings.ReplaceAll(title, "\n", " ")
	title = strings.ReplaceAll(title, "\t", " ")

	title = truncate.Truncate(title, progConfig.HTMLMaxLengthTitle, "...", truncate.PositionEnd)
	htmlHeader := fmt.Sprintf(progConfig.HTMLHeader, title)
	htmlFooter := progConfig.HTMLFooter

	// build html page
	htmlPage := htmlHeader + string(htmlBody) + htmlFooter

	// write html to file
	err = os.WriteFile(destination, []byte(htmlPage), 0666)
	if err != nil {
		fmt.Printf("error [%v] at os.WriteFile()", err)
		return err
	}

	return nil
}
