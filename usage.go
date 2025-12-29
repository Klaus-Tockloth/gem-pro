package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"google.golang.org/genai"
)

/*
printUsage prints the program's usage instructions to the standard output. It displays detailed usage instructions
for the program, including command-line options, examples, and notes about its functionality and terms of service.
*/
func printUsage() {
	fmt.Printf("\nUsage:\n")
	fmt.Printf("  %s [options] [files]\n", progName)

	fmt.Printf("\nExamples [Grouped]:\n")

	fmt.Printf("[Interactive]\n")
	fmt.Printf("  # Interactive mode (type into terminal, file or browser)\n")
	fmt.Printf("  %s\n", progName)

	fmt.Printf("\n[Piping & Output Control]\n")
	fmt.Printf("  # Read from pipe and define specific output filenames\n")
	fmt.Printf("  cat prompt.txt | %s -out my-result\n", progName)
	fmt.Printf("  # Result: my-result.md, my-result.html, my-result.ansi\n")

	fmt.Printf("\n[Working with Local Files]\n")
	fmt.Printf("  # Analyze specific files directly (sent with prompt)\n")
	fmt.Printf("  %s -pro main.go utils.go\n", progName)
	fmt.Printf("  # Use a file list for batch processing context\n")
	fmt.Printf("  %s -filelist source_files.txt\n", progName)

	fmt.Printf("\n[Tools & Grounding]\n")
	fmt.Printf("  # Google Search: Research current events\n")
	fmt.Printf("  %s -google-search\n", progName)
	fmt.Printf("  # URL Context: Summarize a specific webpage\n")
	fmt.Printf("  %s -url-context https://example.com/article\n", progName)
	fmt.Printf("  # Google Maps: Find locations\n")
	fmt.Printf("  %s -google-maps\n", progName)
	fmt.Printf("  # Code Execution: Solve complex math or logic\n")
	fmt.Printf("  %s -code-execution\n", progName)

	fmt.Printf("\n[Chat Mode]\n")
	fmt.Printf("  # Start a conversation session (remember history)\n")
	fmt.Printf("  %s -chatmode -pro\n", progName)
	fmt.Printf("  # Chat with initial file context (files sent only once)\n")
	fmt.Printf("  %s -chatmode config.yaml\n", progName)

	fmt.Printf("\n[Advanced Context: Caching (High Performance / Cost Saving)]\n")
	fmt.Printf("  # Step 1: Create a cache from a library of PDF/Text files\n")
	fmt.Printf("  %s -create-cache *.pdf\n", progName)
	fmt.Printf("  # Step 2: Query the model using the cached content\n")
	fmt.Printf("  %s -include-cache\n", progName)
	fmt.Printf("  # Manage cache\n")
	fmt.Printf("  %s -list-cache\n", progName)
	fmt.Printf("  %s -delete-cache\n", progName)

	fmt.Printf("\n[Advanced Context: RAG (FileSearchStores / Persistent Knowledge)]\n")
	fmt.Printf("  # Step A: Create store\n")
	fmt.Printf("  %s -create-store \"ProjectDocs\"\n", progName)
	fmt.Printf("  # Step B: Add knowledge base documents\n")
	fmt.Printf("  %s -add-to-store \"stores/<ID>\" docs/*.md\n", progName)
	fmt.Printf("  # Step C: Query using the knowledge base\n")
	fmt.Printf("  %s -include-store \"stores/<ID>\"\n", progName)

	groups := []struct {
		name  string
		flags []string
	}{
		{"Model Selection", []string{"lite", "flash", "pro", "flash-image", "pro-image", "default", "list-models"}},
		{"Generation Parameters", []string{"candidates"}},
		{"Grounding / Tools", []string{"code-execution", "google-search", "url-context", "google-maps"}},
		{"Chat & Interaction", []string{"chatmode", "config", "filelist"}},
		{"Output Control", []string{"out"}},
		{"Context: Caching (High Performance)", []string{"create-cache", "include-cache", "list-cache", "delete-cache"}},
		{"Context: Google File Store (Temporary)", []string{"upload-files", "include-files", "list-files", "delete-files"}},
		{"Context: RAG / FileSearchStores (Persistent)", []string{"list-stores", "create-store", "delete-store", "add-to-store", "delete-from-store", "include-store", "list-store-content"}},
	}

	fmt.Printf("\nOptions [Grouped]:\n")
	for _, group := range groups {
		fmt.Printf("\n[%s]\n", group.name)
		for _, flagName := range group.flags {
			flagFlag := flag.Lookup(flagName)
			if flagFlag == nil {
				continue
			}

			var placeholder string
			getter, ok := flagFlag.Value.(interface{ Get() interface{} })
			if ok {
				switch getter.Get().(type) {
				case string, *stringArray, stringArray, []string:
					placeholder = " <string>"
				case int, int32, int64:
					placeholder = " <int>"
				case float32, float64:
					placeholder = " <float>"
				}
			}

			const indent = "\n                          "
			usageText := strings.ReplaceAll(flagFlag.Usage, "\n", indent)
			fmt.Printf("  -%-22s %s", flagFlag.Name+placeholder, usageText)
			fmt.Printf(" (default: %s)", flagFlag.DefValue)
			fmt.Printf("\n")
		}
	}

	var help = `
Notes:
  - Integrate Gemini AI responses into your workflow by prompting via this tool.
  - Submit prompts via the following input channels: Terminal, File, localhost.
  - Output is available in Markdown, HTML, or ANSI format.
  - Files specified on the command line or via the -filelist option are sent to Gemini AI as part of the prompt context.

Notes Concerning Non-Chat Mode (Default):
  - Each prompt is treated independently.
  - The AI does not retain conversation history from previous interactions.
  - Files are sent with every prompt submitted in this mode.

Notes Concerning Chat Mode (-chatmode flag):
  - The AI maintains conversation history within the current session.
  - Files are sent only once, with the initial prompt of the session.
 
Terms of Service for Google Gemini AI:
  Your use of the Google Gemini AI service is subject to the Google Terms of Service (policies.google.com/terms) 
  and the Generative AI Prohibited Use Policy (policies.google.com/terms/generative-ai/use-policy). 
  Visit the Gemini Apps Privacy Hub (support.google.com/gemini?p=privacy_help) to learn how Google uses 
  your Gemini Apps data. See also the Gemini Apps FAQ (gemini.google.com/faq).

Using the Free or Experimental Version of Google Gemini AI:
  - Google may use all input data provided to the free/experimental service to improve Gemini AI.
  - Therefore, do not submit any private or confidential information when using this version.

Using the Paid Version of Google Gemini AI:
  - Google does not use input data provided via the paid service to improve Gemini AI.
  - Connecting your API key to a Google Cloud Platform (GCP) billing account enables access to paid service features.

Required Setup:
  - Obtain a personal Gemini API Key from Google AI Studio (ai.google.dev).
  - Configure the API Key in your shell environment:
    macOS, Linux   : export GEMINI_API_KEY="Your-API-Key"
    Windows (cmd)  : set GEMINI_API_KEY=Your-API-Key
    Windows (Pwsh) : $env:GEMINI_API_KEY="Your-API-Key"
  - Optional: Associate your Gemini API Key with a GCP billing account for paid usage.
 
Tip:
  The included 'prompt-input.html' file provides a basic web interface for crafting prompts 
  and sending them to 'gem-pro' (requires localhost input configuration). 
  A web browser is also helpful for rendering HTML-formatted output from the tool.
`

	fmt.Printf("%s\n", help)
	os.Exit(1)
}

/*
showAvailableGeminiModels retrieves and displays a list of available Gemini AI models from the Gemini API.
It connects to the Gemini API, retrieves a list of available AI models, and prints their details to the
console, including model names, descriptions, and token limits.
*/
func showAvailableGeminiModels(terminalWidth int) {
	// create AI client
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  progConfig.GeminiAPIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		fmt.Printf("error [%v] creating AI client\n", err)
		os.Exit(1)
	}

	// list models
	page, err := client.Models.List(ctx, nil)
	if err != nil {
		fmt.Printf("error [%v] at client.Models.List()\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nGemini AI models:\n")
	for _, modelInfo := range page.Items {
		fmt.Printf("\nName             : %v\n", modelInfo.Name)
		fmt.Printf("DisplayName      : %v\n", modelInfo.DisplayName)
		fmt.Printf("Description      : %v\n", wrapString(modelInfo.Description, terminalWidth, 19))
		fmt.Printf("Version          : %v\n", modelInfo.Version)
		// Endpoints
		// Labels
		// TunedModelInfo
		fmt.Printf("InputTokenLimit  : %v\n", modelInfo.InputTokenLimit)
		fmt.Printf("OutputTokenLimit : %v\n", modelInfo.OutputTokenLimit)
		fmt.Printf("SupportedActions : %v\n", strings.Join(modelInfo.SupportedActions, ", "))
		// DefaultCheckpointID
		// Checkpoints
		fmt.Printf("Temperature      : %v\n", modelInfo.Temperature)
		fmt.Printf("MaxTemperature   : %v\n", modelInfo.MaxTemperature)
		fmt.Printf("TopP             : %v\n", modelInfo.TopP)
		fmt.Printf("TopK             : %v\n", modelInfo.TopK)
		fmt.Printf("Thinking         : %t\n", modelInfo.Thinking)
	}
	fmt.Printf("\n")
}
