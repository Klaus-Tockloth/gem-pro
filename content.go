package main

import (
	"os"

	"google.golang.org/genai"
)

/*
convertFileToContent converts a file into Gemini genai.Content format. It reads a file from the given path
and converts its content into a `genai.Content` object, automatically detecting the MIME type.
*/
func convertFileToContent(filepath string) (*genai.Content, error) {
	mimeType, err := getFileMimeType(filepath)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	return genai.NewContentFromBytes(data, mimeType, "user"), nil
}
