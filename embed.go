package main

import (
	_ "embed"
	"log"
	"os"
)

//go:embed gem-pro.yaml
var gemProYaml []byte

/*
writeConfig writes the embedded default configuration (gemProYaml) to a file named 'gem-pro.yaml'. It
embeds the default configuration from 'gemProYaml' and writes it to a file in the current directory,
used if no config file exists.
*/
func writeConfig() {
	filename := "gem-pro.yaml"
	err := os.WriteFile(filename, gemProYaml, 0666)
	if err != nil {
		log.Fatalf("embed: error [%v] at os.WriteFile(), file = [%s]", err, filename)
	}
}

//go:embed prompt-input.html
var geminiPromptInputHTML []byte

/*
writePromptInput writes the embedded HTML content for prompt input (geminiPromptInputHTML) to a file. It
embeds the HTML for prompt input from 'geminiPromptInputHTML' and writes it to 'prompt-input.html' in the
current directory, providing a default HTML input page.
*/
func writePromptInput() {
	filename := "prompt-input.html"
	err := os.WriteFile(filename, geminiPromptInputHTML, 0666)
	if err != nil {
		log.Fatalf("embed: error [%v] at os.WriteFile(), file = [%s]", err, filename)
	}
}

//go:embed assets/gemini-prompt.css
var assetsGeminiPromptCSS []byte

//go:embed assets/gemini-prompt-303030.svg
var assetsGeminiPrompt303030Svg []byte

//go:embed assets/gemini-prompt-ebebeb.svg
var assetsGeminiPromptEbebebSvg []byte

//go:embed assets/copy-to-clipboard.js
var assetsCopyToClipboardJs []byte

/*
writeAssets writes HTML assets (CSS, SVG, JS files) required for HTML rendering to a specified base path.
It embeds necessary HTML assets like CSS, SVG images, and JavaScript files and writes them to the assets
subdirectory within the given base path.
*/
func writeAssets(basepath string) {
	filename := basepath + "/assets/gemini-prompt.css"
	err := os.WriteFile(filename, assetsGeminiPromptCSS, 0666)
	if err != nil {
		log.Fatalf("embed: error [%v] at os.WriteFile(), file = [%s]", err, filename)
	}

	filename = basepath + "/assets/gemini-prompt-303030.svg"
	err = os.WriteFile(filename, assetsGeminiPrompt303030Svg, 0666)
	if err != nil {
		log.Fatalf("embed: error [%v] at os.WriteFile(), file = [%s]", err, filename)
	}

	filename = basepath + "/assets/gemini-prompt-ebebeb.svg"
	err = os.WriteFile(filename, assetsGeminiPromptEbebebSvg, 0666)
	if err != nil {
		log.Fatalf("embed: error [%v] at os.WriteFile(), file = [%s]", err, filename)
	}

	filename = basepath + "/assets/copy-to-clipboard.js"
	err = os.WriteFile(filename, assetsCopyToClipboardJs, 0666)
	if err != nil {
		log.Fatalf("embed: error [%v] at os.WriteFile(), file = [%s]", err, filename)
	}
}
