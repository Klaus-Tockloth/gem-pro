package main

import (
	"context"
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"google.golang.org/genai"
)

/*
createAIModelSpecificCache creates a new AI model specific cache from given files,
including System Instructions and Tools into the cached content object.
*/
func createAIModelSpecificCache(filesToUpload []FileToHandle) {
	if len(filesToUpload) == 0 {
		fmt.Printf("  nothing to do, no files to upload\n")
		return
	}

	// create AI client
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  progConfig.GeminiAPIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatalf("error [%v] creating Gemini AI client", err)
	}

	// retrieve all cache objects
	for item, err := range client.Caches.All(ctx) {
		if err != nil {
			log.Fatalf("error [%v] retrieving cached content resources from Gemini AI", err)
		}
		if item.DisplayName == progConfig.GeminiCacheName {
			log.Fatalf("error creating AI model specific cache, cache '%s' already exists", progConfig.GeminiCacheName)
		}
	}

	// iterate over all files to upload
	parts := []*genai.Part{}
	for _, fileToUpload := range filesToUpload {
		if fileToUpload.State != "ok" {
			continue
		}

		cacheToHandle.FilesTokenized = append(cacheToHandle.FilesTokenized, fileToUpload)

		// convert file to content
		content, err := convertFileToContent(fileToUpload.Filepath)
		if err != nil {
			fmt.Printf("error [%v] converting file to content\n", err)
			continue
		}
		parts = append(parts, content.Parts[0])
	}

	// prepare system instruction for cache
	var sysInstruction *genai.Content
	if progConfig.SystemInstructionFile != "" {
		sysInstructionBytes, err := os.ReadFile(progConfig.SystemInstructionFile)
		if err != nil {
			fmt.Printf("error [%v] reading system instruction file [%s]\n", err, progConfig.SystemInstructionFile)
			os.Exit(1)
		}
		finalSystemInstruction = string(sysInstructionBytes)
	} else {
		finalSystemInstruction = ""
	}
	if finalSystemInstruction != "" {
		sysInstruction = genai.NewContentFromText(finalSystemInstruction, "user")
	}

	// prepare tools for cache
	tools := []*genai.Tool{}
	if progConfig.GeminiGroundingWithCodeExecution {
		tools = append(tools, &genai.Tool{CodeExecution: &genai.ToolCodeExecution{}})
	}
	if progConfig.GeminiGroundingWithGoogleSearch {
		tools = append(tools, &genai.Tool{GoogleSearch: &genai.GoogleSearch{}})
	}
	if progConfig.GeminiGroundingWithURLContext {
		tools = append(tools, &genai.Tool{URLContext: &genai.URLContext{}})
	}
	if progConfig.GeminiGroundingWithGoogleMaps {
		tools = append(tools, &genai.Tool{GoogleMaps: &genai.GoogleMaps{}})
	}
	if len(includeStores) > 0 {
		tools = append(tools, &genai.Tool{
			FileSearch: &genai.FileSearch{
				FileSearchStoreNames: includeStores,
			},
		})
	}

	// create cached content including SystemInstruction and Tools
	createConfig := &genai.CreateCachedContentConfig{
		TTL:               time.Duration(progConfig.GeminiCacheTimeToLive) * time.Hour,
		DisplayName:       progConfig.GeminiCacheName,
		Contents:          []*genai.Content{{Role: "user", Parts: parts}},
		SystemInstruction: sysInstruction,
	}
	if len(tools) > 0 {
		createConfig.Tools = tools
	}

	cachedContent, err := client.Caches.Create(ctx, progConfig.GeminiAiModel, createConfig)
	if err != nil {
		log.Fatalf("error [%v] creating Gemini AI cache", err)
	}

	// add cached content details
	cacheToHandle.CachedContent = *cachedContent

	cacheToHandle.SystemInstruction = finalSystemInstruction

	var toolNames []string
	if progConfig.GeminiGroundingWithCodeExecution {
		toolNames = append(toolNames, "CodeExecution")
	}
	if progConfig.GeminiGroundingWithGoogleSearch {
		toolNames = append(toolNames, "GoogleSearch")
	}
	if progConfig.GeminiGroundingWithURLContext {
		toolNames = append(toolNames, "URLContext")
	}
	if progConfig.GeminiGroundingWithGoogleMaps {
		toolNames = append(toolNames, "GoogleMaps")
	}
	if len(includeStores) > 0 {
		toolNames = append(toolNames, "FileSearchStores")
	}
	cacheToHandle.Tools = toolNames
}

/*
updateAIModelSpecificCache updates the TTL of the AI model specific cache.
*/
func updateAIModelSpecificCache(ttlHours int) {
	// create AI client
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  progConfig.GeminiAPIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatalf("error [%v] creating Gemini AI client", err)
	}

	aiModel := filepath.Base(progConfig.GeminiAiModel)
	cacheFound := false

	// retrieve all caches
	for item, err := range client.Caches.All(ctx) {
		if err != nil {
			log.Fatalf("error [%v] retrieving cached content resources from Gemini AI", err)
		}

		// AI model specific cache
		if item.DisplayName == progConfig.GeminiCacheName && filepath.Base(item.Model) == aiModel {
			// update cache
			_, err = client.Caches.Update(ctx, item.Name, &genai.UpdateCachedContentConfig{
				TTL: time.Duration(ttlHours) * time.Hour,
			})
			if err != nil {
				log.Fatalf("error [%v] updating cache from Gemini AI", err)
			}
			cacheFound = true
			break
		}
	}

	if !cacheFound {
		fmt.Printf("  error: no AI model specific cache found to update\n")
		os.Exit(1)
	}
}

/*
deleteAIModelSpecificCache deletes AI model specific cache.
*/
func deleteAIModelSpecificCache() {
	// create AI client
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  progConfig.GeminiAPIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatalf("error [%v] creating Gemini AI client", err)
	}

	aiModel := filepath.Base(progConfig.GeminiAiModel)

	// retrieve all caches
	for item, err := range client.Caches.All(ctx) {
		if err != nil {
			log.Fatalf("error [%v] retrieving cached content resources from Gemini AI", err)
		}

		// AI model specific cache
		if item.DisplayName == progConfig.GeminiCacheName && filepath.Base(item.Model) == aiModel {
			// delete cache
			_, err = client.Caches.Delete(ctx, item.Name, &genai.DeleteCachedContentConfig{})
			if err != nil {
				log.Fatalf("error [%v] deleting cache from Gemini AI", err)
			}
			break
		}
	}
}

/*
listAIModelSpecificCache lists AI model specific cache including cached System Instructions and Tools.
Cache details and tokenized files are formatted in Markdown.
*/
func listAIModelSpecificCache(indent string) (string, string) {
	cacheName := ""
	cacheDetails := ""

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
	aiModel := filepath.Base(progConfig.GeminiAiModel)
	cacheFound := false

	// retrieve all caches
	for item, err := range client.Caches.All(ctx) {
		if err != nil {
			log.Fatalf("error [%v] retrieving cached content resources from Gemini AI", err)
		}

		// AI model specific cache
		if item.DisplayName == progConfig.GeminiCacheName && filepath.Base(item.Model) == aiModel {
			// print details of cached content in Markdown format
			cacheDetails += fmt.Sprintf("%s* **Name**: `%s`\n", indent, item.Name)
			cacheDetails += fmt.Sprintf("%s* **DisplayName**: `%s`\n", indent, item.DisplayName)
			cacheDetails += fmt.Sprintf("%s* **Model**: `%s`\n", indent, item.Model)
			cacheDetails += fmt.Sprintf("%s* **CreateTime**: %s\n", indent, item.CreateTime.Local().Format(time.RFC850))

			diff := item.ExpireTime.Sub(now)
			diffInHours := diff.Hours()
			cacheDetails += fmt.Sprintf("%s* **ExpireTime**: %s (%.1f h)\n", indent, item.ExpireTime.Local().Format(time.RFC850), diffInHours)

			if item.UsageMetadata != nil {
				if item.UsageMetadata.AudioDurationSeconds > 0 {
					cacheDetails += fmt.Sprintf("%s* **AudioDuration**: %d (sec)\n", indent, item.UsageMetadata.AudioDurationSeconds)
				}
				if item.UsageMetadata.VideoDurationSeconds > 0 {
					cacheDetails += fmt.Sprintf("%s* **VideoDuration**: %d (sec)\n", indent, item.UsageMetadata.VideoDurationSeconds)
				}
				if item.UsageMetadata.ImageCount > 0 {
					cacheDetails += fmt.Sprintf("%s* **ImageCount**: %d\n", indent, item.UsageMetadata.ImageCount)
				}
				if item.UsageMetadata.TextCount > 0 {
					cacheDetails += fmt.Sprintf("%s* **TextCount**: %d\n", indent, item.UsageMetadata.TextCount)
				}
				if item.UsageMetadata.TotalTokenCount > 0 {
					cacheDetails += fmt.Sprintf("%s* **TotalToken**: %d\n", indent, item.UsageMetadata.TotalTokenCount)
				}
			}
			cacheName = item.Name
			cacheFound = true
			break
		}
	}

	if !cacheFound {
		cacheDetails += fmt.Sprintf("%sno AI model specific cache found\n", indent)
		return cacheName, cacheDetails
	}

	// load saved cache details
	filename := progConfig.GeminiCacheName + "." + filepath.Base(progConfig.GeminiAiModel) + ".gob"
	savedCacheDetails, err := loadCacheDetailsFromFile(filename)
	if err != nil {
		log.Fatalf("error [%v] at loadCacheDetailsFromFile()", err)
	}

	// populate global cacheToHandle
	cacheToHandle = savedCacheDetails

	// verify cache name and AI model
	if savedCacheDetails.CachedContent.Name != cacheName || filepath.Base(savedCacheDetails.CachedContent.Model) != aiModel {
		cacheDetails += fmt.Sprintf("%swarning: unexpected content in file [%s]\n", indent, filename)
	} else {
		// pass system instruction from cache to global variable for prompt rendering
		finalSystemInstruction = savedCacheDetails.SystemInstruction

		// iterate over all tokenized files in cache details and render as Markdown table
		if len(savedCacheDetails.FilesTokenized) > 0 {
			cacheDetails += fmt.Sprintf("\n%s| Path | Size | MIME | Modified |\n", indent)
			cacheDetails += fmt.Sprintf("%s| :--- | :--- | :--- | :--- |\n", indent)
			for _, fileTokenized := range savedCacheDetails.FilesTokenized {
				cacheDetails += fmt.Sprintf("%s| %s | %s | %s | %s |\n", indent,
					fileTokenized.Filepath, fileTokenized.FileSize, fileTokenized.MimeType, fileTokenized.LastUpdate)
			}
		}
	}

	return cacheName, cacheDetails
}

/*
saveCacheDetailsToFile saves AI model specific cache data to file.
*/
func saveCacheDetailsToFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	encoder := gob.NewEncoder(file)
	return encoder.Encode(cacheToHandle)
}

/*
loadCacheDetailsFromFile loads AI model specific cache data from file.
*/
func loadCacheDetailsFromFile(filename string) (CacheToHandle, error) {
	var cacheToHandle CacheToHandle
	file, err := os.Open(filename)
	if err != nil {
		return cacheToHandle, err
	}
	defer func() { _ = file.Close() }()
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&cacheToHandle)
	return cacheToHandle, err
}
