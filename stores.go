package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/genai"
)

/*
listGeminiFileSearchStores lists all available FileSearchStores.
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

	found := false
	for store, err := range client.FileSearchStores.All(ctx) {
		if err != nil {
			log.Fatalf("error [%v] retrieving FileSearchStores", err)
		}
		if !found {
			fmt.Printf("%s%-30s %-30s %-20s\n", indent, "Name (ID)", "Display Name", "Create Time")
			found = true
		}
		fmt.Printf("%s%-30s %-30s %-20s\n",
			indent, store.Name, store.DisplayName, store.CreateTime.Local().Format(time.RFC822))
	}

	if !found {
		fmt.Printf("%sno FileSearchStores found\n", indent)
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

		op, err := client.FileSearchStores.UploadToFileSearchStoreFromPath(ctx,
			fileToHandle.Filepath,
			storeName,
			&genai.UploadToFileSearchStoreConfig{
				DisplayName: fileToHandle.Filepath,
				MIMEType:    fileToHandle.MimeType,
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

	fmt.Printf("%sListing documents in store '%s':\n", indent, storeName)

	found := false
	for doc, err := range client.FileSearchStores.Documents.All(ctx, storeName) {
		if err != nil {
			log.Fatalf("error [%v] retrieving documents from store '%s'", err, storeName)
		}

		if !found {
			fmt.Printf("%s%-40s %-30s %-20s\n",
				indent+"  ", "Document Name (ID)", "Display Name", "Create Time")
			found = true
		}

		fmt.Printf("%s%-40s %-30s %-20s\n",
			indent+"  ", doc.Name, doc.DisplayName, doc.CreateTime.Local().Format(time.RFC822))
	}

	if !found {
		fmt.Printf("%sno documents found in this store\n", indent+"  ")
	}
}
