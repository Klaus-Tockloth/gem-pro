package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"google.golang.org/genai"
)

/*
uploadFilesToGemini uploads files to Google file store. Gemini AI can reference an uploaded file via URL.
An uploaded files is stored for a limited amount of time (e.g., 2 days).
*/
func uploadFilesToGemini(filesToUpload []FileToHandle) {
	// create AI client
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  progConfig.GeminiAPIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatalf("error [%v] creating Gemini AI client", err)
	}

	// iterate over all files to upload
	for _, fileToUpload := range filesToUpload {
		if fileToUpload.State != "ok" {
			continue
		}

		uploadFileConfig := genai.UploadFileConfig{}
		// 'Name' may only contain lowercase alphanumeric characters or dashes (-) and cannot begin or
		//  end with a dash. Due to the above limitations, we let Gemini AI generete an unique 'Name'.
		uploadFileConfig.MIMEType = fileToUpload.MimeType
		uploadFileConfig.DisplayName = fileToUpload.Filepath

		fmt.Printf("  %s\n", fileToUpload.Filepath)

		_, err := client.Files.UploadFromPath(ctx, fileToUpload.Filepath, &uploadFileConfig)
		if err != nil {
			log.Fatalf("error [%v] uploading file to Google File Store", err)
		}
	}
}

/*
deleteFilesFromGemini deletes all uploaded files from Google file store.
*/
func deleteFilesFromGemini() {
	// create AI client
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  progConfig.GeminiAPIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatalf("error [%v] creating Gemini AI client", err)
	}

	// iterate over all files uploaded to Gemini AI
	for file, err := range client.Files.All(ctx) {
		if err != nil {
			log.Fatalf("error [%v] iterating over uploaded files", err)
		}

		fmt.Printf("  %s\n", file.DisplayName)

		_, err := client.Files.Delete(ctx, file.Name, nil)
		if err != nil {
			log.Fatalf("error [%v] deleting file from Google File Store", err)
		}
	}
}

/*
listFilesUploadedToGemini lists all uploaded files in Google file store.
*/
func listFilesUploadedToGemini(indent string) string {
	filelist := ""

	// create AI client
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  progConfig.GeminiAPIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatalf("error [%v] creating Gemini AI client", err)
	}

	now := time.Now()

	// iterate over all files uploaded to Gemini AI
	printHeader := true
	for file, err := range client.Files.All(ctx) {
		if err != nil {
			log.Fatalf("error [%v] iterating over uploaded files", err)
		}
		sizeBytes := -1
		if file.SizeBytes != nil {
			sizeBytes = int(*file.SizeBytes)
		}
		diff := file.ExpirationTime.Sub(now)
		diffInHours := diff.Hours()

		if printHeader {
			filelist += fmt.Sprintf("%sDisplayName, Size, MIMEType, State, RemainingTime\n", indent)
			printHeader = false
		}
		filelist += fmt.Sprintf("%s%s, %.1f (KiB), %s, %s, %.1f (h)\n", indent,
			file.DisplayName, float64(sizeBytes)/1024.0, file.MIMEType, file.State, diffInHours)
	}

	return filelist
}

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
