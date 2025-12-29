package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/genai"
)

/*
listGeminiFileSearchStores lists all available FileSearchStores in a detailed block format.
*/
func listGeminiFileSearchStores(indent string) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  progConfig.GeminiAPIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatalf("error [%v] creating Gemini AI client", err)
	}

	fmt.Printf("\n%sFull Store Name (ID)\n", indent)
	fmt.Printf("%s  Display Name\n", indent)
	fmt.Printf("%s  Create Time, Update Time\n", indent)
	fmt.Printf("%s  Active Docs, Total Size\n\n", indent)

	found := false
	for store, err := range client.FileSearchStores.All(ctx) {
		if err != nil {
			log.Fatalf("error [%v] retrieving FileSearchStores", err)
		}

		found = true
		createTimeStr := store.CreateTime.Local().Format("20060102-150405")
		updateTimeStr := store.UpdateTime.Local().Format("20060102-150405")
		sizeKiB := float64(store.SizeBytes) / 1024.0

		// block format
		fmt.Printf("%s%s\n", indent, store.Name)
		fmt.Printf("%s  %s\n", indent, store.DisplayName)
		fmt.Printf("%s  %s, %s\n", indent, createTimeStr, updateTimeStr)
		fmt.Printf("%s  %d active documents, %.1f KiB\n\n", indent, store.ActiveDocumentsCount, sizeKiB)
	}

	if !found {
		fmt.Printf("%sno FileSearchStores found\n", indent+"  ")
	}
}

/*
createGeminiFileSearchStore creates a new FileSearchStore.
*/
func createGeminiFileSearchStore(displayName string) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx,
		&genai.ClientConfig{
			APIKey:  progConfig.GeminiAPIKey,
			Backend: genai.BackendGeminiAPI,
		})
	if err != nil {
		log.Fatalf("error [%v] creating Gemini AI client", err)
	}

	store, err := client.FileSearchStores.Create(ctx,
		&genai.CreateFileSearchStoreConfig{
			DisplayName: displayName,
		})
	if err != nil {
		log.Fatalf("error [%v] creating FileSearchStore", err)
	}

	fmt.Printf("  Store created successfully:\n")
	fmt.Printf("  Name: %s\n  Display: %s\n", store.Name, store.DisplayName)
}

/*
deleteGeminiFileSearchStore deletes a store by its name (ID).
*/
func deleteGeminiFileSearchStore(storeName string) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  progConfig.GeminiAPIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatalf("error [%v] creating Gemini AI client", err)
	}

	// Force=true deletes the store even if it still contains files.
	forceDelete := true
	err = client.FileSearchStores.Delete(ctx, storeName, &genai.DeleteFileSearchStoreConfig{
		Force: &forceDelete,
	})
	if err != nil {
		log.Fatalf("error [%v] deleting FileSearchStore '%s'", err, storeName)
	}

	fmt.Printf("  Store '%s' deleted successfully.\n", storeName)
}

/*
addFilesToFileSearchStore uploads local files to an existing store
and waits for indexing to complete (polling).
*/
func addFilesToFileSearchStore(storeName string, files []FileToHandle) {
	if len(files) == 0 {
		fmt.Printf("  nothing to do, no files to upload\n")
		return
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  progConfig.GeminiAPIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatalf("error [%v] creating Gemini AI client", err)
	}

	for _, fileToHandle := range files {
		if fileToHandle.State != "ok" {
			continue
		}

		fmt.Printf("  Uploading %s ... ", fileToHandle.Filepath)

		mimeType := fileToHandle.MimeType
		if ReplacementMIMETypeMap != nil {
			replacement, ok := ReplacementMIMETypeMap[mimeType]
			if ok {
				mimeType = replacement
			}
		}

		op, err := client.FileSearchStores.UploadToFileSearchStoreFromPath(ctx,
			fileToHandle.Filepath,
			storeName,
			&genai.UploadToFileSearchStoreConfig{
				DisplayName: fileToHandle.Filepath,
				MIMEType:    mimeType,
			})
		if err != nil {
			fmt.Printf("FAILED: %v\n", err)
			continue
		}

		// Poll for the status of the upload operation.
		for {
			op, err = client.Operations.GetUploadToFileSearchStoreOperation(ctx, op, nil)
			if err != nil {
				fmt.Printf("Error polling status: %v\n", err)
				break
			}
			if op.Done {
				if op.Error != nil {
					fmt.Printf("ERROR: %v\n", op.Error)
				} else {
					fmt.Printf("DONE\n")
				}
				break
			}
			time.Sleep(1 * time.Second)
		}
	}
}

/*
deleteFileFromFileSearchStore deletes a specific document from a FileSearchStore.
documentName is the full resource name (ID) of the document to delete.
*/
func deleteFileFromFileSearchStore(documentName string) {
	if documentName == "" {
		return
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  progConfig.GeminiAPIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatalf("error [%v] creating Gemini AI client", err)
	}

	fmt.Printf("  Deleting document '%s' ... ", documentName)
	forceDelete := true
	err = client.FileSearchStores.Documents.Delete(ctx, documentName, &genai.DeleteDocumentConfig{
		Force: &forceDelete,
	})
	if err != nil {
		fmt.Printf("FAILED: %v\n", err)
	} else {
		fmt.Printf("DONE\n")
	}
}

/*
listFilesInFileSearchStore lists all documents (files) within a specific store.
storeName is the ID/name of the store (e.g., "fileSearchStores/12345").
*/
func listFilesInFileSearchStore(storeName string, indent string) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  progConfig.GeminiAPIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatalf("error [%v] creating Gemini AI client", err)
	}

	fmt.Printf("\nListing documents in store '%s':\n\n", storeName)
	fmt.Printf("%sFull Document Name (ID)\n", indent)
	fmt.Printf("%s  Full Display Name\n", indent)
	fmt.Printf("%s  Create Time, Size, MIMEType\n\n", indent)

	found := false
	for doc, err := range client.FileSearchStores.Documents.All(ctx, storeName) {
		if err != nil {
			log.Fatalf("error [%v] retrieving documents from store '%s'", err, storeName)
		}

		found = true
		createTimeStr := doc.CreateTime.Local().Format("20060102-150405")
		sizeKiB := float64(doc.SizeBytes) / 1024.0

		// block format
		fmt.Printf("%s%s\n", indent, doc.Name)
		fmt.Printf("%s  %s\n", indent, doc.DisplayName)
		fmt.Printf("%s  %s, %.1f KiB, %s\n\n", indent, createTimeStr, sizeKiB, doc.MIMEType)
	}

	if !found {
		fmt.Printf("%sno documents found in this store\n", indent+"  ")
	}
}
