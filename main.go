/*
Purpose:
- gemini prompt (gem-pro)

Description:
- Prompt Google Gemini AI and display the response.

Releases:
  - v0.1.0 - 2025/03/11: initial release
  - v0.2.0 - 2025/03/15: 'GroundingChunks' added to response output
  - v0.3.0 - 2025/03/24: image support added, libs updated, SIGSEGV in main() and processResponse() fixed

Copyright:
- © 2025 | Klaus Tockloth

License:
- MIT License

Contact:
- klaus.tockloth@googlemail.com

Remarks:
- none

Links:
- https://pkg.go.dev/google.golang.org/genai
*/
package main

import (
	"context"
	"flag"
	"fmt"
	"math"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
	"golang.org/x/term"
	"google.golang.org/genai"
)

// general program info
var (
	progName    = strings.TrimSuffix(filepath.Base(os.Args[0]), filepath.Ext(filepath.Base(os.Args[0])))
	progVersion = "v0.3.0"
	progDate    = "2025/03/24"
	progPurpose = "gemini prompt"
	progInfo    = "Prompt Google Gemini AI and display the response."
)

// processing timestamp
var (
	startProcessing  time.Time
	finishProcessing time.Time
)

// markdown to html parser
var markdownParser goldmark.Markdown

// FileToUpload represents all files to uploaded to Gemini
type FileToUpload struct {
	state        string
	filepath     string
	lastUpdate   string
	fileSize     string
	mimeType     string
	errorMessage string
}

// filesToUpload holds list of files to upload to Gemini
var filesToUpload []FileToUpload

/*
main starts this program. It is the entry point of the application, responsible for parsing command-line
arguments, loading configuration, initializing resources, and running the main prompt processing loop.
*/
func main() {
	var err error

	fmt.Printf("\nProgram:\n")
	fmt.Printf("  Name    : %s\n", progName)
	fmt.Printf("  Release : %s - %s\n", progVersion, progDate)
	fmt.Printf("  Purpose : %s\n", progPurpose)
	fmt.Printf("  Info    : %s\n", progInfo)

	// request terminal width
	terminalWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		terminalWidth = 132
	}

	// command line parameter
	candidates, temperature, topp, topk, maxtokens, uploads, config, models := processCommandLineFlags()

	if !fileExists(*config) {
		writeConfig()
	}
	if !fileExists("./assets") {
		err = os.Mkdir("./assets", 0750)
		if err != nil && !os.IsExist(err) {
			fmt.Printf("error [%v] at os.Mkdir()\n", err)
			os.Exit(1)
		}
		writeAssets(".")
	}

	if !fileExists("./prompt-input.html") {
		writePromptInput()
	}

	err = loadConfiguration(*config)
	if err != nil {
		fmt.Printf("error [%v] loading configuration\n", err)
		os.Exit(1)
	}

	if *models {
		showAvailableGeminiModels(terminalWidth)
		os.Exit(1)
	}

	// build list of files given via command line
	filesToUpload = buildGivenFiles(flag.Args(), *uploads)

	// shows files given via command line
	fmt.Printf("\nFiles given via command line:\n")
	if len(filesToUpload) == 0 {
		fmt.Printf("  none\n")
	} else {
		for _, fileToUpload := range filesToUpload {
			if fileToUpload.state != "error" {
				fmt.Printf("  %-5s %s (%s, %s, %s)\n",
					fileToUpload.state, fileToUpload.filepath,
					fileToUpload.lastUpdate, fileToUpload.fileSize, fileToUpload.mimeType)
			} else {
				fmt.Printf("  %-5s %s %s\n",
					fileToUpload.state, fileToUpload.filepath, fileToUpload.errorMessage)
			}
		}
	}

	// show configuration
	showConfiguration()

	// initialize this program
	initializeProgram()

	// overwrite YAML config values with cli parameters
	overwriteConfigValues(candidates, temperature, topp, topk, maxtokens)

	// create markdown parser (WithUnsafe() ensures to render potentially dangerous links like "file:///Users/...")
	markdownParser = goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithRendererOptions(html.WithUnsafe()),
	)

	// create AI client
	ctx := context.Background()
	var client *genai.Client
	if progConfig.GeneralInternetProxy != "" {
		// create a ProxyRoundTripper with API key and proxy URL from configuration
		proxyRoundTripper := &ProxyRoundTripper{
			APIKey:   progConfig.GeminiAPIKey,
			ProxyURL: progConfig.GeneralInternetProxy,
		}

		// create a custom HTTP client using the ProxyRoundTripper
		httpClient := &http.Client{
			Transport: proxyRoundTripper,
		}

		// APIKey shouldn't be necessary because the key is set in ProxyRoundTripper but without
		// the option, NewClient() attempts to authenticate via Google Cloud SDK (ADC)

		// indirect internet connection: client -> proxy -> internet
		client, err = genai.NewClient(ctx, &genai.ClientConfig{
			APIKey:     progConfig.GeminiAPIKey,
			HTTPClient: httpClient,
			Backend:    genai.BackendGeminiAPI,
		})
	} else {
		// direct internet connection: client -> internet
		client, err = genai.NewClient(ctx, &genai.ClientConfig{
			APIKey:  progConfig.GeminiAPIKey,
			Backend: genai.BackendGeminiAPI,
		})
	}
	if err != nil {
		fmt.Printf("error [%v] creating AI client\n", err)
		os.Exit(1)
	}

	// get and print Gemini AI model information
	geminiModelInfo, err := client.Models.Get(ctx, progConfig.GeminiAiModel, nil)
	if err != nil {
		fmt.Printf("error [%v] getting AI model information\n", err)
		return
	}
	printGeminiModelInfo(geminiModelInfo, terminalWidth)

	// generate and print Gemini model configuration
	geminiModelConfig := generateGeminiModelConfig()
	printGeminiModelConfig(geminiModelConfig, terminalWidth)

	// define prompt channel
	promptChannel := make(chan string)

	// set up signal handling for shutdown (e.g. Ctrl-C)
	shutdownTrigger := make(chan os.Signal, 1)
	signal.Notify(shutdownTrigger, syscall.SIGINT)  // kill -SIGINT pid -> interrupt
	signal.Notify(shutdownTrigger, syscall.SIGTERM) // kill -SIGTERM pid -> terminated

	fmt.Printf("\nProgram termination:\n")
	fmt.Printf("  Press CTRL-C to terminate this program.\n\n")

	// start graceful shutdown handler
	go handleShutdown(shutdownTrigger)

	// start input readers
	inputPossibilities := startInputReaders(promptChannel, progConfig)

	// main loop: 'Prompt Google Gemini AI'
	for {
		fmt.Printf("Waiting for input from %s ...\n", strings.Join(inputPossibilities, ", "))

		// read prompt from channel
		prompt := <-promptChannel
		prompt = strings.TrimSpace(prompt)

		now := time.Now()
		if progConfig.NotifyPrompt {
			err = runCommand(progConfig.NotifyPromptApplication)
			if err != nil {
				fmt.Printf("error [%v] at runCommand()\n", err)
			}
		}

		// build prompt parts (filedata, text prompt)
		contents := []*genai.Content{}
		for _, fileToUpload := range filesToUpload {
			if fileToUpload.state == "error" {
				continue
			}
			// convert file to content
			content, err := convertFileToContent(fileToUpload.filepath)
			if err != nil {
				fmt.Printf("error [%v] converting file to content\n", err)
				continue
			}
			contents = append(contents, content)
		}
		contents = append(contents, genai.NewUserContentFromText(prompt))

		fmt.Printf("%02d:%02d:%02d: Processing prompt ...\n", now.Hour(), now.Minute(), now.Second())
		processPrompt(prompt)

		dumpDataToFile(os.O_TRUNC|os.O_WRONLY, "gemini model config", geminiModelConfig)
		dumpDataToFile(os.O_APPEND|os.O_CREATE|os.O_WRONLY, "gemini prompt contents", contents)

		// generate content
		startProcessing = time.Now()
		resp, respErr := client.Models.GenerateContent(ctx,
			progConfig.GeminiAiModel,
			contents,
			geminiModelConfig,
		)
		finishProcessing = time.Now()

		dumpDataToFile(os.O_APPEND|os.O_CREATE|os.O_WRONLY, "gemini response", resp)
		dumpDataToFile(os.O_APPEND|os.O_CREATE|os.O_WRONLY, "gemini error", err)

		// trigger response notification
		if progConfig.NotifyResponse {
			err = runCommand(progConfig.NotifyResponseApplication)
			if err != nil {
				fmt.Printf("error [%v] at runCommand()\n", err)
			}
		}

		// handle response
		handleResponse(resp, respErr, prompt)
	}
}

/*
processCommandLineFlags parses command-line flags and returns pointers to their values. It uses the flag package
to define and parse command-line flags, making configuration options available when the program is run.
*/
func processCommandLineFlags() (*int, *float64, *float64, *int, *int, *string, *string, *bool) {
	candidates := flag.Int("candidates", -1, "specifies number of AI responses (overwrites YAML config)")
	temperature := flag.Float64("temperature", -1.0, "specifies variation range of AI responses (overwrites YAML config)")
	topp := flag.Float64("topp", -1.0, "maximum cumulative probability of tokens to consider when sampling (overwrites YAML config)")
	topk := flag.Int("topk", -1, "maximum number of tokens to consider when sampling (overwrites YAML config)")
	maxtokens := flag.Int("maxtokens", -1, "max output tokens (useful to force short content, overwrites YAML config)")
	uploads := flag.String("uploads", "", "name of list with files to upload to AI (one file per line)")
	dir, _ := filepath.Split(os.Args[0])
	defaultConfigFile := dir + progName + ".yaml"
	config := flag.String("config", defaultConfigFile, "name of YAML config file")
	models := flag.Bool("models", false, "show all AI Gemini models and terminate")

	flag.Usage = printUsage
	flag.Parse()

	return candidates, temperature, topp, topk, maxtokens, uploads, config, models
}

/*
overwriteConfigValues overwrites configuration values in `progConfig` with values provided via command-line flags.
It updates the `progConfig` struct with values from command-line flags, allowing users to override settings from
the YAML configuration file.
*/
func overwriteConfigValues(candidates *int, temperature *float64, topp *float64, topk *int, maxtokens *int) {
	// overwrite YAML config values with cli parameters
	if *candidates > 0 {
		progConfig.GeminiCandidateCount = int32(*candidates)
	}
	if *temperature > -1.0 {
		progConfig.GeminiTemperature = float32(*temperature)
	}
	if *topp > -1.0 {
		progConfig.GeminiTopP = float32(*topp)
	}
	if *topk > -1 {
		progConfig.GeminiTopK = float32(*topk)
	}
	if *maxtokens > 0 {
		progConfig.GeminiMaxOutputTokens = int32(*maxtokens)
	}
}

/*
handleResponse processes the response received from the Gemini AI model. It manages the AI response, including
error handling, output formatting, saving history, and triggering output applications for different formats
like Markdown and HTML.
*/
func handleResponse(resp *genai.GenerateContentResponse, respErr error, prompt string) {
	now := finishProcessing
	fmt.Printf("%02d:%02d:%02d: Processing response ...\n", now.Hour(), now.Minute(), now.Second())
	if respErr != nil {
		processError(respErr)
	} else {
		processResponse(resp)
	}

	// print prompt and response to terminal
	if progConfig.AnsiOutput {
		printPromptResponseToTerminal()
	}

	// copy ansi file to history
	if progConfig.AnsiHistory {
		ansiDestinationFile := buildDestinationFilename(now, prompt, progConfig.HistoryFilenameExtensionAnsi)
		ansiDestinationPathFile := filepath.Join(progConfig.AnsiHistoryDirectory, ansiDestinationFile)
		copyFile(progConfig.AnsiPromptResponseFile, ansiDestinationPathFile)
	}

	// markdown prompt and response file: nothing to do
	commandLine := fmt.Sprintf(progConfig.MarkdownOutputApplication, progConfig.MarkdownPromptResponseFile)

	// copy markdown file to history
	if progConfig.MarkdownHistory {
		markdownDestinationFile := buildDestinationFilename(now, prompt, progConfig.HistoryFilenameExtensionMarkdown)
		markdownDestinationPathFile := filepath.Join(progConfig.MarkdownHistoryDirectory, markdownDestinationFile)
		copyFile(progConfig.MarkdownPromptResponseFile, markdownDestinationPathFile)
		commandLine = fmt.Sprintf(progConfig.MarkdownOutputApplication, "\""+markdownDestinationPathFile+"\"")
	}

	// open markdown document in application
	if progConfig.MarkdownOutput {
		err := runCommand(commandLine)
		if err != nil {
			fmt.Printf("error [%v] at runCommand()\n", err)
		}
	}

	// build prompt and response html page
	commandLine = fmt.Sprintf(progConfig.HTMLOutputApplication, progConfig.HTMLPromptResponseFile)
	_ = buildHTMLPage(prompt, progConfig.HTMLPromptResponseFile, progConfig.HTMLPromptResponseFile)

	// copy html file to history
	if progConfig.HTMLHistory {
		htmlDestinationFile := buildDestinationFilename(now, prompt, progConfig.HistoryFilenameExtensionHTML)
		htmlDestinationPathFile := filepath.Join(progConfig.HTMLHistoryDirectory, htmlDestinationFile)
		copyFile(progConfig.HTMLPromptResponseFile, htmlDestinationPathFile)
		commandLine = fmt.Sprintf(progConfig.HTMLOutputApplication, "\""+htmlDestinationPathFile+"\"")
	}

	// open html page in application
	if progConfig.HTMLOutput {
		err := runCommand(commandLine)
		if err != nil {
			fmt.Printf("error [%v] at runCommand()\n", err)
		}
	}
}

/*
printGeminiModelInfo prints detailed information about a Gemini AI model to the console. It retrieves and
formats detailed information of a given Gemini AI model and prints it to the console, including token
limits, version, and description.
*/
func printGeminiModelInfo(geminiModelInfo *genai.Model, terminalWidth int) {
	// calculate words from tokens
	inputTokenLimitWordsLower := float64(geminiModelInfo.InputTokenLimit) / 100.0 * 60.0
	inputTokenLimitWordsLower = math.Floor(inputTokenLimitWordsLower/100.0) * 100.0
	inputTokenLimitWordsUpper := float64(geminiModelInfo.InputTokenLimit) / 100.0 * 80.0
	inputTokenLimitWordsUpper = math.Floor(inputTokenLimitWordsUpper/100.0) * 100.0
	outputTokenLimitWordsLower := float64(geminiModelInfo.OutputTokenLimit) / 100.0 * 60.0
	outputTokenLimitWordsLower = math.Floor(outputTokenLimitWordsLower/100.0) * 100.0
	outputTokenLimitWordsUpper := float64(geminiModelInfo.OutputTokenLimit) / 100.0 * 80.0
	outputTokenLimitWordsUpper = math.Floor(outputTokenLimitWordsUpper/100.0) * 100.0

	fmt.Printf("\nGemini model information:\n")
	fmt.Printf("  Name              : %v\n", geminiModelInfo.Name)
	fmt.Printf("  DisplayName       : %v\n", geminiModelInfo.DisplayName)
	fmt.Printf("  Description       : %v\n", wrapString(geminiModelInfo.Description, terminalWidth, 22))
	fmt.Printf("  Version           : %v\n", geminiModelInfo.Version)
	// Endpoints
	// Labels
	// TunedModeInfo
	fmt.Printf("  InputTokenLimit   : %v (approx. %.0f-%.0f english words)\n", geminiModelInfo.InputTokenLimit, inputTokenLimitWordsLower, inputTokenLimitWordsUpper)
	fmt.Printf("  OutputTokenLimit  : %v (approx. %.0f-%.0f english words)\n", geminiModelInfo.OutputTokenLimit, outputTokenLimitWordsLower, outputTokenLimitWordsUpper)
	fmt.Printf("  SupportedActions  : %v\n", strings.Join(geminiModelInfo.SupportedActions, ", "))
}

/*
handleShutdown handles program termination signals (SIGINT and SIGTERM). It listens for shutdown signals
and performs a graceful program exit when a signal is received, ensuring resources are properly released.
*/
func handleShutdown(shutdownTrigger chan os.Signal) {
	<-shutdownTrigger
	fmt.Printf("\nShutdown signal received. Exiting gracefully ...\n")
	os.Exit(0)
}

/*
startInputReaders initializes and starts input reader goroutines based on the program configuration. It sets
up and starts goroutines for reading prompts from different input sources like terminal, file, or localhost,
based on the configuration.
*/
func startInputReaders(promptChannel chan string, config ProgConfig) []string {
	inputPossibilities := []string{}

	// input from keyboard
	if config.InputFromTerminal {
		go readPromptFromKeyboard(promptChannel)
		inputPossibilities = append(inputPossibilities, "Terminal")
	}

	// input from file
	if config.InputFromFile {
		if !fileExists(config.InputFile) {
			file, err := os.Create(config.InputFile)
			if err != nil {
				fmt.Printf("error [%v] creating input prompt text file\n", err)
				return inputPossibilities
			}
			file.Close()
		}
		go readPromptFromFile(config.InputFile, promptChannel)
		inputPossibilities = append(inputPossibilities, "File")
	}

	// input from localhost
	if config.InputFromLocalhost {
		addr := fmt.Sprintf("localhost:%d", config.InputLocalhostPort)
		go func() {
			http.HandleFunc("/", readPromptFromLocalhost(promptChannel))
			err := http.ListenAndServe(addr, nil)
			if err != nil {
				fmt.Printf("error [%v] starting internal webserver\n", err)
				return
			}
		}()
		inputPossibilities = append(inputPossibilities, addr)
	}

	return inputPossibilities
}

/*
buildGivenFiles builds a list of files provided via command-line (list, args). It processes file paths from
command-line arguments and a file list, checks their state, and prepares a list of FileToUpload structures
for further processing.
*/
func buildGivenFiles(args []string, uploads string) []FileToUpload {
	var filesFromList []string
	var err error

	if uploads != "" {
		filesFromList, err = slurpFile(uploads)
		if err != nil {
			fmt.Printf("error [%v] reading list of files to upload to AI\n", err)
		}
	}

	files := filesFromList
	files = append(files, args...)

	for _, file := range files {
		fileToUpload := FileToUpload{filepath: file}

		fileInfo, err := os.Stat(file)
		if err != nil {
			fileToUpload.state = "error"
			fileToUpload.errorMessage = fmt.Sprintf("error [%v] at os.Stat()", err)
		} else {
			mimeType, err := getFileMimeType(file)
			info := "ok"
			if err != nil {
				info = "error"
			}
			if mimeType == "application/octet-stream" {
				info = "warn"
			}
			if err == nil {
				fileToUpload.state = info
				fileToUpload.fileSize = fmt.Sprintf("%.1f KiB", float64(fileInfo.Size())/1024.0)
				fileToUpload.lastUpdate = fileInfo.ModTime().Format("20060102-150405")
				fileToUpload.mimeType = mimeType
			} else {
				fileToUpload.state = info
				fileToUpload.errorMessage = fmt.Sprintf("error [%v] at getFileMimeType()", err)
			}
		}
		filesToUpload = append(filesToUpload, fileToUpload)
	}

	return filesToUpload
}
