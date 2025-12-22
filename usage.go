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

	fmt.Printf("\nExamples:\n")
	fmt.Printf("  # Basic: Start and type prompt into terminal, file, browser\n")
	fmt.Printf("  %s\n", progName)
	fmt.Printf("\n  # Prompt via Pipe\n")
	fmt.Printf("  echo \"Explain this code\" | %s main.go\n", progName)
	fmt.Printf("\n  # Using a file list for context\n")
	fmt.Printf("  %s -filelist sources.txt\n", progName)
	fmt.Printf("\n  # RAG Workflow: Create store and add files\n")
	fmt.Printf("  %s -create-store \"MyProject\"\n", progName)
	fmt.Printf("  %s -add-to-store \"stores/123\" *.go\n", progName)

	// TODO: remove
	fmt.Printf("\nOptions:\n")
	flag.PrintDefaults()

	groups := []struct {
		name  string
		flags []string
	}{
		{"Model Selection", []string{"lite", "flash", "pro", "default", "list-models"}},
		{"Generation Parameters", []string{"candidates", "temperature", "topp"}},
		{"Grounding", []string{"code-execution", "google-search", "google-maps"}},
		{"RAG / File Search Stores", []string{"list-stores", "create-store", "add-to-store", "list-store-content", "delete-store", "include-store"}},
		{"Context Caching", []string{"create-cache", "include-cache", "list-cache", "delete-cache"}},
		{"Google File Store", []string{"upload-files", "include-files", "list-files", "delete-files"}},
		{"Input & Configuration", []string{"config", "filelist", "chatmode"}},
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
Remark Concerning Options:
  A default value of -1 for numeric options indicates that the option was not set via the command line. 
  The program will use the value from the YAML configuration file or the API's default if not specified there.

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
