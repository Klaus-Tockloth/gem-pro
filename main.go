//go:fmt off
/*
Purpose:
- gemini prompt (gem-pro)

Description:
- Prompts Google Gemini AI and displays the response.

Releases:
  - v0.1.0 - 2025-03-11: initial release
  - v0.2.0 - 2025-03-15: 'GroundingChunks' added to response output
  - v0.3.0 - 2025-03-24: image support added, libs updated, SIGSEGV in main() and processResponse() fixed
  - v0.3.1 - 2025-03-28: libs updated, clean up markdown data given by Gemini
  - v0.3.2 - 2025-03-28: web search sources as numbered list, clean up markdown data revised
  - v0.4.0 - 2025-04-02: libs updated, chat mode feature added, compiled with go v1.24.2
  - v0.4.1 - 2025-04-05: user info concerning prompt processing mode (chat, non-chat)
  - v0.5.0 - 2025-06-18: file and cache support, thoughts support, libs updated, go v1.24.4, options added
  - v0.5.1 - 2025-07-05: control output ThinkingBudget, libs updated, CSS improved
  - v0.5.2 - 2025-07-13: CSS improved, markdown cleanup before parsing, libs updated, compiled with go v1.24.5
  - v0.5.3 - 2025-07-30: libs updated, fix for panic: 'interface conversion: ast.Node is *ast.AutoLink, not *ast.Link'
  - v0.6.0 - 2025-09-17: 'code execution' support added, 'Tool use tokens' added to 'Tokens' details,
                         thinking config handling modified, option '-replace-mimetypes' added,
						 libs updated, compiled with go v1.25.1
  - v0.7.0 - 2025-09-20: various options added, default for 'code execution' changed, libs updated
  - v0.8.0 - 2025-10-01: documentation improved, temperature+topp for technical prompts, libs updated
  - v0.9.0 - 2025-10-11: yaml config enriched (latest models, Google Maps), libs updated, compiled with go v1.25.2
  - v0.9.1 - 2025-10-21: segmentation violation in 'output.go' fixed, libs updated, compiled with go v1.25.3

Copyright:
- © 2025 | Klaus Tockloth

License:
- MIT License

Contact:
- klaus.tockloth@googlemail.com

Remarks:
- none

ToDos:
- Support grounding references in response (e.g., "... lorem ipsum.[7][8]" and later "7. Webpage XY").
- Support for "Function Calling" (e.g. 'dynamic' functions via 'yaegi').
- Support for "Stop Sequence".

Links:
- https://pkg.go.dev/google.golang.org/genai
*/
//go:fmt on

// main package
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
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
	progVersion = "v0.9.1"
	progDate    = "2025-10-21"
	progPurpose = "gemini prompt"
	progInfo    = "Prompts Google Gemini AI and displays the response."
)

// processing timestamp
var (
	startProcessing  time.Time
	finishProcessing time.Time
)

// markdown to html parser
var markdownParser goldmark.Markdown

// FileToHandle represents all files to handle in prompt or Gemini context
type FileToHandle struct {
	State        string
	Filepath     string
	LastUpdate   string
	FileSize     string
	MimeType     string
	ErrorMessage string
}

// filesToHandle holds list of files to handle in prompt or Gemini context
var filesToHandle []FileToHandle

// CacheToHandle represents all tokenized files in AI model specific cache
type CacheToHandle struct {
	CachedContent  genai.CachedContent
	FilesTokenized []FileToHandle
}

// cacheToHandle holds all data concerning AI model specific cache
var cacheToHandle CacheToHandle

// ReplacementMIMETypeMap holds two MIME types for replacing 'key' with 'value'
var ReplacementMIMETypeMap map[string]string

// command line parameters
var (
	liteModel        = flag.Bool("lite", false, "Specifies the Gemini AI lite model to use.")
	flashModel       = flag.Bool("flash", false, "Specifies the Gemini AI flash model to use.")
	proModel         = flag.Bool("pro", false, "Specifies the Gemini AI pro model to use.")
	defaultModel     = flag.Bool("default", false, "Specifies the Gemini AI default model to use.")
	candidates       = flag.Int("candidates", -1, "Specifies the number of candidate responses the AI should generate.\nOverrides the value in the YAML config.")
	temperature      = flag.Float64("temperature", -1.0, "Controls the randomness of the AI's responses. Higher values (e.g., 1.8) increase creativity/diversity;\nlower values increase focus/determinism. Overrides the value in the YAML config.")
	topp             = flag.Float64("topp", -1.0, "Sets the cumulative probability threshold for token selection during sampling (Top-P / nucleus sampling).\nOverrides the value in the YAML config.")
	topk             = flag.Int("topk", -1, "Sets the maximum number of tokens to consider at each sampling step (Top-K sampling).\nOverrides the value in the YAML config.")
	maxtokens        = flag.Int("maxtokens", -1, "Sets the maximum number of tokens for the generated response. Useful for constraining output length.\nOverrides the value in the YAML config.")
	filelist         = flag.String("filelist", "", "Specifies a file containing a list of files to upload (one filename per line).\nThese files will be included with the prompt(s).")
	config           = flag.String("config", progName+".yaml", "Specifies the name of the YAML configuration file.")
	listModels       = flag.Bool("list-models", false, "Lists all available Gemini AI models and exits.")
	chatmode         = flag.Bool("chatmode", false, "Enables chat mode, where the AI remembers conversation history within a session.")
	uploadFiles      = flag.Bool("upload-files", false, "Uploads given files to Google File Store and exits.")
	deleteFiles      = flag.Bool("delete-files", false, "Deletes given files from Google File Store and exits.")
	listFiles        = flag.Bool("list-files", false, "Lists given files in Google File Store and exits.")
	includeFiles     = flag.Bool("include-files", false, "Includes all uploaded files from Google File Store in prompt to Gemini AI.")
	createCache      = flag.Bool("create-cache", false, "Creates a new AI model specific cache from given files and exits.")
	deleteCache      = flag.Bool("delete-cache", false, "Deletes AI model specific cache and exits.")
	listCache        = flag.Bool("list-cache", false, "Lists AI model specific cache and exits.")
	includeCache     = flag.Bool("include-cache", false, "Includes AI model specific cache in prompt to Gemini AI.")
	replaceMIMETypes = flag.String("replace-mimetypes", "", "Replaces MIME type 'A' with 'B' (e.g. 'text/x-perl=text/plain').\nExpects a comma separates key-value list of MIME type pairs.")
	thinkingBudget   = flag.Int("thinking-budget", -1, "Sets the maximum thinking budget.\nThe value '0' disables thinking mode (not possible for all models).")
	codeExecution    = flag.Bool("code-execution", false, "Lets Gemini use code to solve complex tasks.")
	googleSearch     = flag.Bool("google-search", false, "Grounding with Google Search.")
)

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

	flag.Usage = printUsage
	flag.Parse()

	// track which flags were actually set by the user
	setFlags := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) {
		setFlags[f.Name] = true
	})

	// build replacement MIME type map
	if *replaceMIMETypes != "" {
		ReplacementMIMETypeMap, err = parseMIMETypeList(*replaceMIMETypes)
		if err != nil {
			fmt.Printf("error [%v] parsing replacement MIME type list\n", err)
			os.Exit(1)
		}
	}

	if !fileExists(*config) {
		writeConfig()
	}

	// 'assets' in current directory (to render current HTML file in current directory)
	directory := "./assets"
	if !dirExists(directory) {
		err = os.Mkdir(directory, 0750)
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

	// set Gemini AI model
	progConfig.GeminiAiModel = progConfig.GeminiDefaultAiModel
	switch {
	case *liteModel:
		progConfig.GeminiAiModel = progConfig.GeminiLiteAiModel
	case *flashModel:
		progConfig.GeminiAiModel = progConfig.GeminiFlashAiModel
	case *proModel:
		progConfig.GeminiAiModel = progConfig.GeminiProAiModel
	case *defaultModel:
		progConfig.GeminiAiModel = progConfig.GeminiDefaultAiModel
	}
	if progConfig.GeminiAiModel == "" {
		fmt.Printf("empty Gemini AI model not allowed\n")
		os.Exit(1)
	}

	// build list of files given via command line
	filesToHandle = buildGivenFiles(flag.Args(), *filelist)

	// shows files given via command line
	fmt.Printf("\nFiles given via command line:\n")
	if len(filesToHandle) == 0 {
		fmt.Printf("  none\n")
	} else {
		for _, fileToHandle := range filesToHandle {
			if fileToHandle.State != "error" {
				// add replacement MIME type (e.g. 'text/x-perl -> text/plain')
				mimeType := fileToHandle.MimeType
				if ReplacementMIMETypeMap != nil {
					replacement, ok := ReplacementMIMETypeMap[fileToHandle.MimeType]
					if ok {
						mimeType += fmt.Sprintf(" -> %s", replacement)
					}
				}
				fmt.Printf("  %-5s %s (%s, %s, %s)\n",
					fileToHandle.State, fileToHandle.Filepath, fileToHandle.LastUpdate, fileToHandle.FileSize, mimeType)
			} else {
				fmt.Printf("  %-5s %s %s\n",
					fileToHandle.State, fileToHandle.Filepath, fileToHandle.ErrorMessage)
			}
		}
	}

	// handle standalone actions
	handleStandaloneFileActions()
	handleStandaloneCacheActions()

	if *listModels {
		showAvailableGeminiModels(terminalWidth)
		os.Exit(0)
	}

	if *includeFiles {
		filelist := listFilesUploadedToGemini("  ")
		fmt.Printf("\nInclude files given via Google file store:\n")
		if len(filelist) == 0 {
			fmt.Printf("  none\n")
		} else {
			fmt.Printf("%s", filelist)
		}
	}

	cacheName := ""
	cacheDetails := ""
	if *includeCache {
		cacheName, cacheDetails = listAIModelSpecificCache("  ")
		fmt.Printf("\nInclude AI model specific cache:\n")
		if len(cacheName) == 0 {
			fmt.Printf("  error: no AI model specific cache found\n\n")
			os.Exit(1)
		}
		fmt.Printf("%s", cacheDetails)
	}

	// show configuration
	showConfiguration()

	// initialize this program
	initializeProgram()

	// overwrite YAML config values with cli parameters
	overwriteConfigValues(setFlags)

	// create markdown parser (WithUnsafe() ensures to render potentially dangerous links like "file:///Users/...")
	markdownParser = goldmark.New(
		goldmark.WithExtensions(extension.GFM, &TargetBlankExtension{}),
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

	// generate and print Gemini model configuration (adds cache if defined)
	geminiModelConfig := generateGeminiModelConfig(cacheName)
	printGeminiModelConfig(geminiModelConfig, terminalWidth)

	// define prompt channel
	promptChannel := make(chan string)

	// set up signal handling for shutdown (e.g. Ctrl-C)
	shutdownTrigger := make(chan os.Signal, 1)
	signal.Notify(shutdownTrigger, syscall.SIGINT)  // kill -SIGINT pid -> interrupt
	signal.Notify(shutdownTrigger, syscall.SIGTERM) // kill -SIGTERM pid -> terminated

	fmt.Printf("\nOperation mode:\n")
	if *chatmode {
		fmt.Printf("  Running in chat mode.\n")
	} else {
		fmt.Printf("  Running in non-chat mode.\n")
	}

	fmt.Printf("\nProgram termination:\n")
	fmt.Printf("  Press CTRL-C to terminate this program.\n\n")

	// start graceful shutdown handler
	go handleShutdown(shutdownTrigger)

	// start input readers
	inputPossibilities := startInputReaders(promptChannel, progConfig)

	// create chat mode session
	chat := &genai.Chat{}
	chatNumber := 1
	if *chatmode {
		chat, err = client.Chats.Create(ctx, progConfig.GeminiAiModel, geminiModelConfig, nil)
		if err != nil {
			fmt.Printf("error [%v] creating Gemini chat mode session\n", err)
			os.Exit(1)
		}
	}

	// start main loop: Prompt Gemini AI
	// ---------------------------------
	var resp *genai.GenerateContentResponse
	var respErr error
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

		contents := []*genai.Content{} // prompt in non-chat mode
		parts := []genai.Part{}        // prompt in chat mode

		// build prompt parts (filedata, text prompt) of type '[]*genai.Content' for non-chat mode
		if !*chatmode {
			// handle files from commandline
			for _, fileToHandle := range filesToHandle {
				if fileToHandle.State == "error" {
					continue
				}
				// convert file to content
				content, err := convertFileToContent(fileToHandle.Filepath)
				if err != nil {
					fmt.Printf("error [%v] converting file to content\n", err)
					continue
				}
				contents = append(contents, content)
			}
			if *includeFiles {
				// handle uploaded files from Google file store
				for file, err := range client.Files.All(ctx) {
					if err != nil {
						log.Fatalf("error [%v] iterating over uploaded files", err)
					}
					contents = append(contents, genai.NewContentFromURI(file.URI, file.MIMEType, "user"))
				}
			}
			// add text prompt
			contents = append(contents, genai.NewContentFromText(prompt, "user"))
		}

		// build prompt parts (filedata, uploaded files, text prompt) of type '[]genai.Part' for chat mode
		if *chatmode {
			// in chat mode we only add filedata to initial chat prompt
			if chatNumber == 1 {
				// handle files from commandline
				for _, fileToHandle := range filesToHandle {
					if fileToHandle.State == "error" {
						continue
					}
					// convert file to content
					content, err := convertFileToContent(fileToHandle.Filepath)
					if err != nil {
						fmt.Printf("error [%v] converting file to content\n", err)
						continue
					}
					parts = append(parts, *content.Parts[0])
				}
				if *includeFiles {
					// handle uploaded files from Google file store
					for file, err := range client.Files.All(ctx) {
						if err != nil {
							log.Fatalf("error [%v] iterating over uploaded files", err)
						}
						parts = append(parts, *genai.NewPartFromFile(*file))
					}
				}
			}
			parts = append(parts, *genai.NewPartFromText(prompt))
		}

		if *chatmode {
			fmt.Printf("%02d:%02d:%02d: Processing prompt in chat mode ...\n", now.Hour(), now.Minute(), now.Second())
		} else {
			fmt.Printf("%02d:%02d:%02d: Processing prompt in non-chat mode ...\n", now.Hour(), now.Minute(), now.Second())
		}

		processPrompt(prompt, *chatmode, chatNumber)

		dumpDataToFile(os.O_TRUNC|os.O_WRONLY, "gemini model config", geminiModelConfig)
		dumpDataToFile(os.O_APPEND|os.O_CREATE|os.O_WRONLY, "gemini prompt contents", contents)

		// generate content
		startProcessing = time.Now()
		if *chatmode {
			// chat mode
			resp, respErr = chat.SendMessage(ctx, parts...)
		} else {
			// non-chat mode
			resp, respErr = client.Models.GenerateContent(ctx, progConfig.GeminiAiModel, contents, geminiModelConfig)
		}
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

		// increase chat number
		if *chatmode {
			chatNumber++
		}
	}
	// end main loop: Prompt Gemini AI
	// -------------------------------
}

/*
overwriteConfigValues overwrites configuration values in `progConfig` with values provided via command-line flags.
It updates the `progConfig` struct with values from command-line flags, allowing users to override settings from
the YAML configuration file.
*/
func overwriteConfigValues(setFlags map[string]bool) {
	if setFlags["candidates"] {
		progConfig.GeminiCandidateCount = int32(*candidates)
	}
	if setFlags["temperature"] {
		progConfig.GeminiTemperature = float32(*temperature)
	}
	if setFlags["topp"] {
		progConfig.GeminiTopP = float32(*topp)
	}
	if setFlags["topk"] {
		progConfig.GeminiTopK = float32(*topk)
	}
	if setFlags["maxtokens"] {
		progConfig.GeminiMaxOutputTokens = int32(*maxtokens)
	}
	if setFlags["thinking-budget"] {
		progConfig.GeminiMaxThinkingBudget = int32(*thinkingBudget)
	}
	if setFlags["code-execution"] {
		progConfig.GeminiCodeExecution = *codeExecution
	}
	if setFlags["google-search"] {
		progConfig.GeminiGroundigWithGoogleSearch = *googleSearch
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
			_ = file.Close()
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
command-line arguments and a file list, checks their state, and prepares a list of FileToHandle structures
for further processing.
*/
func buildGivenFiles(args []string, filelist string) []FileToHandle {
	var filesFromList []string
	var err error

	if filelist != "" {
		filesFromList, err = slurpFile(filelist)
		if err != nil {
			fmt.Printf("error [%v] reading list of files to upload to AI\n", err)
		}
	}

	files := filesFromList
	files = append(files, args...)

	for _, file := range files {
		fileToHandle := FileToHandle{Filepath: file}

		fileInfo, err := os.Stat(file)
		if err != nil {
			fileToHandle.State = "error"
			fileToHandle.ErrorMessage = fmt.Sprintf("error [%v] at os.Stat()", err)
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
				fileToHandle.State = info
				fileToHandle.FileSize = fmt.Sprintf("%.1f KiB", float64(fileInfo.Size())/1024.0)
				fileToHandle.LastUpdate = fileInfo.ModTime().Format("20060102-150405")
				fileToHandle.MimeType = mimeType
			} else {
				fileToHandle.State = info
				fileToHandle.ErrorMessage = fmt.Sprintf("error [%v] at getFileMimeType()", err)
			}
		}
		filesToHandle = append(filesToHandle, fileToHandle)
	}

	return filesToHandle
}

/*
handleStandaloneFileActions handles file upload, delete and list.
*/
func handleStandaloneFileActions() {
	switch {
	case *uploadFiles:
		fmt.Printf("\nUploading files to Google file store:\n")
		uploadFilesToGemini(filesToHandle)
		fmt.Printf("\n")
		os.Exit(0)

	case *deleteFiles:
		fmt.Printf("\nDeleting files from Google file store:\n")
		deleteFilesFromGemini()
		fmt.Printf("\n")
		os.Exit(0)

	case *listFiles:
		filelist := listFilesUploadedToGemini("  ")
		fmt.Printf("\nFiles in Google file store:\n")
		if len(filelist) == 0 {
			fmt.Printf("  none\n\n")
		} else {
			fmt.Printf("%s\n", filelist)
		}
		os.Exit(0)
	}
}

/*
handleStandaloneCacheActions handles cache create, delete and list.
*/
func handleStandaloneCacheActions() {
	switch {
	case *createCache:
		fmt.Printf("\nCreating AI model specific cache:\n")
		createAIModelSpecificCache(filesToHandle)
		filename := progConfig.GeminiCacheName + "." + filepath.Base(progConfig.GeminiAiModel) + ".gob"
		err := saveCacheDetailsToFile(filename)
		if err != nil {
			fmt.Printf("error [%v] saving cache details to file\n", err)
		}
		_, cacheDetails := listAIModelSpecificCache("  ")
		fmt.Printf("\nAI model specific cache created:\n")
		if len(cacheDetails) == 0 {
			fmt.Printf("  none\n\n")
		} else {
			fmt.Printf("%s\n", cacheDetails)
		}
		os.Exit(0)

	case *deleteCache:
		fmt.Printf("\nDeleting AI model specific cache:\n")
		deleteAIModelSpecificCache()
		_, cacheDetails := listAIModelSpecificCache("  ")
		fmt.Printf("%s\n", cacheDetails)
		os.Exit(0)

	case *listCache:
		_, cacheDetails := listAIModelSpecificCache("  ")
		fmt.Printf("\nListing AI model specific cache:\n")
		if len(cacheDetails) == 0 {
			fmt.Printf("  none\n\n")
		} else {
			fmt.Printf("%s\n", cacheDetails)
		}
		os.Exit(0)
	}
}
