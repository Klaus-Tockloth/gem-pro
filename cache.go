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
createApIModelSpecificCache creates new AI model specific cache from given files.
All files (content parts) thereby form one cache object.
The cache object must consist of at least 1024 or 2048 tokens.
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

	// create cached content
	cachedContent, err := client.Caches.Create(ctx, progConfig.GeminiAiModel, &genai.CreateCachedContentConfig{
		TTL:         time.Duration(progConfig.GeminiCacheTimeToLive) * time.Hour,
		DisplayName: progConfig.GeminiCacheName,
		Contents:    []*genai.Content{{Role: "user", Parts: parts}},
	})
	if err != nil {
		log.Fatalf("error [%v] creating Gemini AI cache", err)
	}

	// add cached content details
	cacheToHandle.CachedContent = *cachedContent
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
listAIModelSpecificCache lists AI model specific cache.
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
			// print details of cached content
			cacheDetails += fmt.Sprintf("%sName          : %s\n", indent, item.Name)
			cacheDetails += fmt.Sprintf("%sDisplayName   : %s\n", indent, item.DisplayName)
			cacheDetails += fmt.Sprintf("%sModel         : %s\n", indent, item.Model)
			cacheDetails += fmt.Sprintf("%sCreateTime    : %s\n", indent, item.CreateTime.Local().Format(time.RFC850))
			// cacheDetails += fmt.Sprintf("%sUpdateTime    : %s\n", indent, item.UpdateTime.Local().Format(time.RFC850))

			diff := item.ExpireTime.Sub(now)
			diffInHours := diff.Hours()
			cacheDetails += fmt.Sprintf("%sExpireTime    : %s (%.1f h)\n", indent, item.ExpireTime.Local().Format(time.RFC850), diffInHours)
			if item.UsageMetadata != nil {
				if item.UsageMetadata.AudioDurationSeconds > 0 {
					cacheDetails += fmt.Sprintf("%sAudioDuration : %d (sec)\n", indent, item.UsageMetadata.AudioDurationSeconds)
				}
				if item.UsageMetadata.VideoDurationSeconds > 0 {
					cacheDetails += fmt.Sprintf("%sVideoDuration : %d (sec)\n", indent, item.UsageMetadata.VideoDurationSeconds)
				}
				if item.UsageMetadata.ImageCount > 0 {
					cacheDetails += fmt.Sprintf("%sImageCount    : %d\n", indent, item.UsageMetadata.ImageCount)
				}
				if item.UsageMetadata.TextCount > 0 {
					cacheDetails += fmt.Sprintf("%sTextCount     : %d\n", indent, item.UsageMetadata.TextCount)
				}
				if item.UsageMetadata.TotalTokenCount > 0 {
					cacheDetails += fmt.Sprintf("%sTotalToken    : %d\n", indent, item.UsageMetadata.TotalTokenCount)
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

	// verify cache name and AI model
	if savedCacheDetails.CachedContent.Name != cacheName || filepath.Base(savedCacheDetails.CachedContent.Model) != aiModel {
		cacheDetails += fmt.Sprintf("%swarning: unexpected content in file [%s]\n", indent, filename)
	} else {
		// iterate over all tokenized files in cache details
		printHeader := true
		for _, fileTokenized := range savedCacheDetails.FilesTokenized {
			if printHeader {
				cacheDetails += fmt.Sprintf("\n%sDisplayName, Size, MIMEType, UpdateTime\n", indent)
				printHeader = false
			}
			cacheDetails += fmt.Sprintf("%s%s, %s, %s, %s\n", indent,
				fileTokenized.Filepath, fileTokenized.FileSize, fileTokenized.MimeType, fileTokenized.LastUpdate)
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
	defer file.Close()
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
	defer file.Close()
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&cacheToHandle)
	return cacheToHandle, err
}
