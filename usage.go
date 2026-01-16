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
printUsage prints the program's usage instructions to the standard output.
*/
func printUsage() {
	fmt.Printf("\nUsage:\n")
	fmt.Printf("  %s [options] [files]\n", progName)

	fmt.Printf("\nExamples:\n")

	// Interactive
	fmt.Printf("  %-30s %s\n", "[Interactive Mode]", progName)

	// Piping
	fmt.Printf("  %-30s %s\n", "[Piped Input]", "cat task.txt | "+progName+" -out result")
	fmt.Printf("  %-30s %s\n", "[Pure Response]", "echo \"Hello\" | "+progName+" -pure-response")

	// Files
	fmt.Printf("  %-30s %s\n", "[Local source files]", progName+" -pro main.go utils.go")
	fmt.Printf("  %-30s %s\n", "[Use list of files]", progName+" -filelist sources.txt")

	// Grounding
	fmt.Printf("  %-30s %s\n", "[Web-Search & No Maps]", progName+" -google-search -google-maps=false")
	fmt.Printf("  %-30s %s\n", "[Research specific URL]", progName+" -url-context https://go.dev")

	// RAG / Stores
	fmt.Printf("  %-30s %s\n", "[Create Knowledge Base]", progName+" -create-store \"LegalDocs\"")
	fmt.Printf("  %-30s %s\n", "[Populate Knowledge Base]", progName+" -add-to-store stores/12345 -filelist docs.txt")
	fmt.Printf("  %-30s %s\n", "[Query with Store ID]", progName+" -include-store stores/12345")

	// Caching
	fmt.Printf("  %-30s %s\n", "[Cache large files]", progName+" -create-cache *.pdf")

	// Groups
	groups := []struct {
		name  string
		flags []string
	}{
		{"Model Selection", []string{"lite", "flash", "pro", "flash-image", "pro-image", "default", "list-models"}},
		{"Generation Parameters", []string{"candidates", "pure-response"}},
		{"Grounding & Tools", []string{"code-execution", "google-search", "url-context", "google-maps"}},
		{"Chat & Interaction", []string{"chatmode", "verbose", "config", "filelist"}},
		{"Output Control", []string{"out"}},
		{"Context: Caching (High Perf)", []string{"create-cache", "include-cache", "list-cache", "delete-cache"}},
		{"Context: Google File Store", []string{"upload-files", "include-files", "list-files", "delete-files"}},
		{"Context: RAG (Persistent)", []string{"list-stores", "create-store", "delete-store", "add-to-store", "delete-from-store", "include-store", "list-store-content"}},
	}

	fmt.Printf("\nOptions:\n")
	fmt.Println("  Note: CLI flags override settings in your YAML configuration. Use '-flag=false' to disable boolean options.")

	for _, group := range groups {
		fmt.Printf("\n  [%s]\n", group.name)
		for _, flagName := range group.flags {
			f := flag.Lookup(flagName)
			if f == nil {
				continue
			}

			// Placeholder
			placeholder := ""
			if getter, ok := f.Value.(interface{ Get() interface{} }); ok {
				switch getter.Get().(type) {
				case string, *stringArray, stringArray, []string:
					placeholder = " <str>"
				case int, int32, int64:
					placeholder = " <int>"
				}
			}

			// Flag Name + Placeholder
			flagPart := fmt.Sprintf("-%s%s", f.Name, placeholder)

			// Indent Usage Text
			usageLines := strings.Split(f.Usage, "\n")
			fmt.Printf("    %-28s %s", flagPart, usageLines[0])

			// Show default if appropriate
			if f.DefValue != "" && f.DefValue != "false" && f.DefValue != "[]" {
				fmt.Printf(" (default: %s)", f.DefValue)
			}
			fmt.Println()

			// Indent further description lines
			for _, line := range usageLines[1:] {
				fmt.Printf("    %-28s %s\n", "", strings.TrimSpace(line))
			}
		}
	}

	fmt.Printf("\nCore Concepts:\n")
	fmt.Printf("  %-30s %s\n", "[Input Channels]", "Interactive Terminal, File-Watch (prompt-input.txt), localhost:4242.")
	fmt.Printf("  %-30s %s\n", "[Terminal Inject]", "Type '<<< filename.txt' in terminal to load file content as prompt.")
	fmt.Printf("  %-30s %s\n", "[Output Formats]", "Markdown (raw), ANSI (terminal color), HTML (browser with JS features).")
	fmt.Printf("  %-30s %s\n", "[Chat Mode]", "AI remembers history. Files are sent only with the FIRST prompt.")
	fmt.Printf("  %-30s %s\n", "[Non-Chat Mode]", "Each prompt is isolated. Files are sent with EVERY prompt.")
	fmt.Printf("  %-30s %s\n", "[Exit Interactive]", "Type Ctrl+C to quit.")

	fmt.Printf("\nEnvironment Variables:\n")
	fmt.Printf("  %-30s %s\n", "[GEMINI_API_KEY]", "Your API Key from ai.google.dev (Mandatory).")
	fmt.Printf("  %-30s %s\n", "[HTTPS_PROXY]", "Used if set and no proxy is defined in YAML.")

	fmt.Printf("\nTerms & Privacy:\n")
	fmt.Printf("  %-30s %s\n", "[Free Tier]", "Google uses your data to improve their models. DO NOT use private data.")
	fmt.Printf("  %-30s %s\n", "[Paid Tier]", "Data is not used for training (requires GCP billing).")

	fmt.Printf("\nFor more details, see the embedded README.md or visit ai.google.dev.\n\n")
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
