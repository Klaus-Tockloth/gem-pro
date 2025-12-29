package main

import (
	"embed"
	"io/fs"
	"log"
	"os"
	"path/filepath"
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
	err := os.WriteFile(filename, gemProYaml, 0600)
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
	err := os.WriteFile(filename, geminiPromptInputHTML, 0600)
	if err != nil {
		log.Fatalf("embed: error [%v] at os.WriteFile(), file = [%s]", err, filename)
	}
}

//go:embed README.md
var readmeBytes []byte

/*
writeReadme writes the embedded README.md to the current directory.
*/
func writeReadme() {
	filename := "README.md"
	err := os.WriteFile(filename, readmeBytes, 0600)
	if err != nil {
		log.Fatalf("embed: error [%v] at os.WriteFile(), file = [%s]", err, filename)
	}
}

//go:embed gem-pro.png
var gemProPngBytes []byte

/*
writeGemProPng writes the embedded gem-pro.png to the current directory.
*/
func writeGemProPng() {
	filename := "gem-pro.png"
	err := os.WriteFile(filename, gemProPngBytes, 0600)
	if err != nil {
		log.Fatalf("embed: error [%v] at os.WriteFile(), file = [%s]", err, filename)
	}
}

//go:embed assets
var assetsFS embed.FS

/*
writeAssets writes all embedded files from the 'assets/' directory
to the provided base path. It iterates over the embedded filesystem
and creates directories and files accordingly.
*/
func writeAssets(basepath string) {
	err := fs.WalkDir(assetsFS, "assets", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		targetPath := filepath.Join(basepath, path)

		if d.IsDir() {
			err := os.MkdirAll(targetPath, 0700)
			if err != nil {
				return err
			}
		} else {
			content, err := assetsFS.ReadFile(path)
			if err != nil {
				return err
			}

			err = os.WriteFile(targetPath, content, 0600)
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		log.Fatalf("embed: error [%v] at writeAssets() exploring assetsFS", err)
	}
}
