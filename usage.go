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
	fmt.Printf("  %s\n", progName)
	fmt.Printf("  %s -candidates 2\n", progName)
	fmt.Printf("  %s -temperature 1.8\n", progName)
	fmt.Printf("  %s *.go README.md\n", progName)
	fmt.Printf("  %s -uploads ganymed-project-files.txt\n", progName)

	fmt.Printf("\nOptions:\n")
	flag.PrintDefaults()

	fmt.Printf("\nRemark concerning Options:\n")
	fmt.Printf("  A default value of -1 only means that the value is not set by the user.\n")

	fmt.Printf("\nNotes:\n")
	fmt.Printf("  - You can prompt Gemini AI and integrate the response into your workflow.\n")
	fmt.Printf("  - You can submit prompts via this input channels: Terminal, File, localhost\n")
	fmt.Printf("  - Output is available in the following formats: Markdown, HTML, ANSI\n")
	fmt.Printf("  - Each prompt is self-contained (chat is not support).\n")
	fmt.Printf("  - Specified files are transmitted to Gemini AI as part of the prompts.\n")

	fmt.Printf("\nDisclaimer:\n")
	fmt.Printf("  This application is for evaluating the concept of integrating and using AI in\n")
	fmt.Printf("  a personalized work environment. All v0.x versions require a Gemini API key,\n")
	fmt.Printf("  enabling free and limited use of 'Google Gemini AI'.\n")
	fmt.Printf("  A Gemini API key is associated with a personal Google account, allowing trial\n")
	fmt.Printf("  use of 'Google Gemini AI'. The Gemini API key is not intended for permanent or\n")
	fmt.Printf("  extensive use. From v1.0 onwards, the application will switch to using a regular\n")
	fmt.Printf("  Google Gemini account.\n")

	fmt.Printf("\nTerms of service apply to 'Google Gemini AI':\n")
	fmt.Printf("  The Google Terms of Service (policies.google.com/terms) and the Generative AI\n")
	fmt.Printf("  Prohibited Use Policy (policies.google.com/terms/generative-ai/use-policy) apply\n")
	fmt.Printf("  to 'Google Gemini AI' service. Visit the Gemini Apps Privacy Hub\n")
	fmt.Printf("  (support.google.com/gemini?p=privacy_help) to learn more about how Google uses\n")
	fmt.Printf("  your Gemini Apps data. See also the Gemini Apps FAQ (gemini.google.com/faq).\n")

	fmt.Printf("\nNotes concerning the freely available version of 'Google Gemini AI'\n")
	fmt.Printf("  - All input data will be used by Google to improve 'Gemini AI'.\n")
	fmt.Printf("  - Therefore, do not process any private or confidential data.\n")

	fmt.Printf("\nRequired:\n")
	fmt.Printf("  - Obtain a personal 'Gemini API Key' from Google.\n")
	fmt.Printf("  - Configure the API Key in your program environment:\n")
	fmt.Printf("    macOS, Linux : export GEMINI_API_KEY=Your-API-Key\n")
	fmt.Printf("    Windows      : set GEMINI_API_KEY Your-API-Key\n")

	fmt.Printf("\nTip:\n")
	fmt.Printf("  In practice, a browser is useful for both creating prompts and presenting\n")
	fmt.Printf("  the output. The simple 'prompt-input.html' webpage can be used for creating\n")
	fmt.Printf("  and sending prompts to 'localhost'.\n")

	fmt.Printf("\n")
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
	}
	fmt.Printf("\n")
}
