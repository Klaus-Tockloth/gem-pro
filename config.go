package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"google.golang.org/genai"
	"gopkg.in/yaml.v3"
)

// ProgConfig represents program configuration
type ProgConfig struct {
	// Gemini configuration
	GeminiAPIKey string `yaml:"GeminiAPIKey"`

	GeminiAiModel           string // one of the following models
	GeminiLiteAiModel       string `yaml:"GeminiLiteAiModel"`
	GeminiFlashAiModel      string `yaml:"GeminiFlashAiModel"`
	GeminiProAiModel        string `yaml:"GeminiProAiModel"`
	GeminiFlashImageAiModel string `yaml:"GeminiFlashImageAiModel"`
	GeminiProImageAiModel   string `yaml:"GeminiProImageAiModel"`
	GeminiDefaultAiModel    string `yaml:"GeminiDefaultAiModel"`

	GeminiResponseModalities []string `yaml:"GeminiResponseModalities"`
	GeminiImageAspectRatio   string   `yaml:"GeminiImageAspectRatio"`
	GeminiImageResolution    string   `yaml:"GeminiImageResolution"`

	GeminiCandidateCount  *int32   `yaml:"GeminiCandidateCount"`
	GeminiPureResponse    bool     `yaml:"GeminiPureResponse"`
	GeminiMaxOutputTokens *int32   `yaml:"GeminiMaxOutputTokens"`
	GeminiTemperature     *float32 `yaml:"GeminiTemperature"`
	GeminiTopP            *float32 `yaml:"GeminiTopP"`
	GeminiTopK            *float32 `yaml:"GeminiTopK"`

	GeminiGroundingWithCodeExecution    bool     `yaml:"GeminiGroundingWithCodeExecution"`
	GeminiGroundingWithGoogleSearch     bool     `yaml:"GeminiGroundingWithGoogleSearch"`
	GeminiGroundingWithURLContext       bool     `yaml:"GeminiGroundingWithURLContext"`
	GeminiGroundigWithGoogleMaps        bool     `yaml:"GeminiGroundigWithGoogleMaps"`
	GeminiGroundingWithFileSearchStores []string `yaml:"GeminiGroundingWithFileSearchStores"`

	GeminiMaxThinkingBudget    *int32 `yaml:"GeminiMaxThinkingBudget"`
	GeminiThinkingLevel        string `yaml:"GeminiThinkingLevel"`
	GeminiIncludeThoughts      bool   `yaml:"GeminiIncludeThoughts"`
	GeminiCacheName            string `yaml:"GeminiCacheName"`
	GeminiCacheTimeToLive      int    `yaml:"GeminiCacheTimeToLive"`
	GeminiInputMediaResolution string `yaml:"GeminiInputMediaResolution"`

	// Markdown configuration
	MarkdownPromptResponseFile       string `yaml:"MarkdownPromptResponseFile"`
	MarkdownOutput                   bool   `yaml:"MarkdownOutput"`
	MarkdownOutputApplication        string
	MarkdownOutputApplicationMacOS   string `yaml:"MarkdownOutputApplicationMacOS"`
	MarkdownOutputApplicationLinux   string `yaml:"MarkdownOutputApplicationLinux"`
	MarkdownOutputApplicationWindows string `yaml:"MarkdownOutputApplicationWindows"`
	MarkdownOutputApplicationOther   string `yaml:"MarkdownOutputApplicationOther"`
	MarkdownHistory                  bool   `yaml:"MarkdownHistory"`
	MarkdownHistoryDirectory         string `yaml:"MarkdownHistoryDirectory"`

	// ANSI configuration
	AnsiRendering          bool   `yaml:"AnsiRendering"`
	AnsiPromptResponseFile string `yaml:"AnsiPromptResponseFile"`
	AnsiOutput             bool   `yaml:"AnsiOutput"`
	AnsiOutputLineLength   int    `yaml:"AnsiOutputLineLength"`
	AnsiHistory            bool   `yaml:"AnsiHistory"`
	AnsiHistoryDirectory   string `yaml:"AnsiHistoryDirectory"`
	AnsiOutputTheme        string `yaml:"AnsiOutputTheme"`

	// HTML configuration
	HTMLRendering                bool   `yaml:"HTMLRendering"`
	HTMLPromptResponseFile       string `yaml:"HTMLPromptResponseFile"`
	HTMLOutput                   bool   `yaml:"HTMLOutput"`
	HTMLOutputApplication        string
	HTMLOutputApplicationMacOS   string              `yaml:"HTMLOutputApplicationMacOS"`
	HTMLOutputApplicationLinux   string              `yaml:"HTMLOutputApplicationLinux"`
	HTMLOutputApplicationWindows string              `yaml:"HTMLOutputApplicationWindows"`
	HTMLOutputApplicationOther   string              `yaml:"HTMLOutputApplicationOther"`
	HTMLHistory                  bool                `yaml:"HTMLHistory"`
	HTMLHistoryDirectory         string              `yaml:"HTMLHistoryDirectory"`
	HTMLMaxLengthTitle           int                 `yaml:"HTMLMaxLengthTitle"`
	HTMLReplaceElements          []map[string]string `yaml:"HTMLReplaceElements"`
	HTMLHeader                   string              `yaml:"HTMLHeader"`
	HTMLFooter                   string              `yaml:"HTMLFooter"`

	// Input configuration
	InputFromTerminal  bool   `yaml:"InputFromTerminal"`
	InputFromFile      bool   `yaml:"InputFromFile"`
	InputFile          string `yaml:"InputFile"`
	InputFromLocalhost bool   `yaml:"InputFromLocalhost"`
	InputLocalhostPort int    `yaml:"InputLocalhostPort"`

	// Notification configuration
	NotifyPrompt                     bool `yaml:"NotifyPrompt"`
	NotifyPromptApplication          string
	NotifyPromptApplicationMacOS     string `yaml:"NotifyPromptApplicationMacOS"`
	NotifyPromptApplicationLinux     string `yaml:"NotifyPromptApplicationLinux"`
	NotifyPromptApplicationWindows   string `yaml:"NotifyPromptApplicationWindows"`
	NotifyPromptApplicationOther     string `yaml:"NotifyPromptApplicationOther"`
	NotifyResponse                   bool   `yaml:"NotifyResponse"`
	NotifyResponseApplication        string
	NotifyResponseApplicationMacOS   string `yaml:"NotifyResponseApplicationMacOS"`
	NotifyResponseApplicationLinux   string `yaml:"NotifyResponseApplicationLinux"`
	NotifyResponseApplicationWindows string `yaml:"NotifyResponseApplicationWindows"`
	NotifyResponseApplicationOther   string `yaml:"NotifyResponseApplicationOther"`

	// General configuration
	GeneralInternetProxy string   `yaml:"GeneralInternetProxy"`
	MIMETypeReplacements []string `yaml:"MIMETypeReplacements"`

	// System instruction
	UserSystemInstruction    string `yaml:"UserSystemInstruction"`
	IncludeSystemInstruction bool   `yaml:"IncludeSystemInstruction"`
}

// progConfig contains program configuration
var progConfig = ProgConfig{}

/*
loadConfiguration loads program configuration from a YAML file. It reads the specified YAML file,
unmarshals it into the global `progConfig` struct, performs extensive validation checks on the loaded
values, sets OS-specific configurations (e.g., application paths), and resolves secrets like the
Gemini API key and proxy credentials using the `getPassword` helper. It returns an error if reading,
unmarshalling, validation, or secret retrieval fails.
*/
func loadConfiguration(configFile string) error {
	operatingSystem := runtime.GOOS

	source, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("error [%w] reading configuration file", err)
	}
	err = yaml.Unmarshal(source, &progConfig)
	if err != nil {
		return fmt.Errorf("error [%w] unmarshalling configuration file", err)
	}

	// gemini
	if progConfig.GeminiAPIKey == "" {
		return fmt.Errorf("empty GeminiAPIKey not allowed")
	}
	if progConfig.GeminiCandidateCount == nil || *progConfig.GeminiCandidateCount <= 0 {
		return fmt.Errorf("empty or invalid GeminiCandidateCount not allowed")
	}

	// markdown
	if progConfig.MarkdownPromptResponseFile == "" {
		return fmt.Errorf("empty MarkdownPromptResponseFile not allowed")
	}
	switch operatingSystem {
	case "darwin":
		progConfig.MarkdownOutputApplication = progConfig.MarkdownOutputApplicationMacOS
	case "linux":
		progConfig.MarkdownOutputApplication = progConfig.MarkdownOutputApplicationLinux
	case "windows":
		progConfig.MarkdownOutputApplication = progConfig.MarkdownOutputApplicationWindows
	default:
		progConfig.MarkdownOutputApplication = progConfig.MarkdownOutputApplicationOther
	}
	if progConfig.MarkdownOutput && progConfig.MarkdownOutputApplication == "" {
		return fmt.Errorf("empty operating system specific MarkdownOutputApplication not allowed")
	}
	if progConfig.MarkdownHistory && progConfig.MarkdownHistoryDirectory == "" {
		return fmt.Errorf("empty MarkdownHistoryDirectory not allowed")
	}

	// ansi
	if progConfig.AnsiRendering && progConfig.AnsiPromptResponseFile == "" {
		return fmt.Errorf("empty AnsiPromptResponseFile not allowed")
	}
	if progConfig.AnsiHistory && progConfig.AnsiHistoryDirectory == "" {
		return fmt.Errorf("empty AnsiHistoryDirectory not allowed")
	}

	// html
	if progConfig.HTMLRendering && progConfig.HTMLPromptResponseFile == "" {
		return fmt.Errorf("empty HTMLPromptResponseFile not allowed")
	}
	switch operatingSystem {
	case "darwin":
		progConfig.HTMLOutputApplication = progConfig.HTMLOutputApplicationMacOS
	case "linux":
		progConfig.HTMLOutputApplication = progConfig.HTMLOutputApplicationLinux
	case "windows":
		progConfig.HTMLOutputApplication = progConfig.HTMLOutputApplicationWindows
	default:
		progConfig.HTMLOutputApplication = progConfig.HTMLOutputApplicationOther
	}
	if progConfig.HTMLOutput && progConfig.HTMLOutputApplication == "" {
		return fmt.Errorf("empty operating system specific HTMLOutputApplication not allowed")
	}
	if progConfig.HTMLHistory && progConfig.HTMLHistoryDirectory == "" {
		return fmt.Errorf("empty HTMLHistoryDirectory not allowed")
	}

	// input
	if progConfig.InputFromFile && progConfig.InputFile == "" {
		return fmt.Errorf("empty InputFile not allowed")
	}

	// notification
	switch operatingSystem {
	case "darwin":
		progConfig.NotifyPromptApplication = progConfig.NotifyPromptApplicationMacOS
	case "linux":
		progConfig.NotifyPromptApplication = progConfig.NotifyPromptApplicationLinux
	case "windows":
		progConfig.NotifyPromptApplication = progConfig.NotifyPromptApplicationWindows
	default:
		progConfig.NotifyPromptApplication = progConfig.NotifyPromptApplicationOther
	}
	if progConfig.NotifyPrompt && progConfig.NotifyPromptApplication == "" {
		return fmt.Errorf("empty operating system specific NotifyPromptApplication not allowed")
	}
	switch operatingSystem {
	case "darwin":
		progConfig.NotifyResponseApplication = progConfig.NotifyResponseApplicationMacOS
	case "linux":
		progConfig.NotifyResponseApplication = progConfig.NotifyResponseApplicationLinux
	case "windows":
		progConfig.NotifyResponseApplication = progConfig.NotifyResponseApplicationWindows
	default:
		progConfig.NotifyResponseApplication = progConfig.NotifyResponseApplicationOther
	}
	if progConfig.NotifyResponse && progConfig.NotifyResponseApplication == "" {
		return fmt.Errorf("empty operating system specific NotifyResponseApplication not allowed")
	}

	// get api-key (password)
	progConfig.GeminiAPIKey, err = getPassword(progConfig.GeminiAPIKey)
	if err != nil {
		return fmt.Errorf("error [%w] getting api-key", err)
	}

	// get internet proxy (password)
	if progConfig.GeneralInternetProxy != "" {
		progConfig.GeneralInternetProxy, err = getPassword(progConfig.GeneralInternetProxy)
		if err != nil {
			return fmt.Errorf("error [%w] getting internet proxy", err)
		}
	}

	// MIME type replacement
	if len(progConfig.MIMETypeReplacements) > 0 {
		mimeMap, err := parseMIMETypeReplacements(progConfig.MIMETypeReplacements)
		if err != nil {
			return fmt.Errorf("error [%s] parsing MIME type replacements", err)
		}
		ReplacementMIMETypeMap = mimeMap
	}

	// thinking
	if progConfig.GeminiMaxThinkingBudget != nil && progConfig.GeminiThinkingLevel != "" {
		return fmt.Errorf("do not set both thinking budget and thinking level")
	}
	if progConfig.GeminiThinkingLevel != "" {
		switch strings.ToLower(progConfig.GeminiThinkingLevel) {
		case "low":
		case "high":
		default:
			return fmt.Errorf("unsupported thinking level [%s]", progConfig.GeminiThinkingLevel)
		}
	}

	// media resolution
	if progConfig.GeminiInputMediaResolution != "" {
		switch strings.ToLower(progConfig.GeminiInputMediaResolution) {
		case "low":
		case "medium":
		case "high":
		default:
			return fmt.Errorf("unsupported media resolution [%s]", progConfig.GeminiInputMediaResolution)
		}
	}

	// image generation
	if progConfig.GeminiImageAspectRatio != "" {
		validRatios := map[string]bool{"auto": true, "1:1": true,
			"9:16": true, "16:9": true,
			"3:4": true, "4:3": true,
			"2:3": true, "3:2": true,
			"4:5": true, "5:4": true,
			"21:9": true,
		}
		if !validRatios[progConfig.GeminiImageAspectRatio] {
			fmt.Printf("warning: unusual aspect ratio [%s]\n", progConfig.GeminiImageAspectRatio)
		}
	}

	return nil
}

/*
showConfiguration shows / prints the loaded program configuration to the console. It displays the current
program configuration settings to the user in the console for review.
*/
func showConfiguration() {
	// general notes
	fmt.Printf("\nNotes concerning the freely available version of 'Google Gemini AI':\n")
	fmt.Printf("  See the help page for the 'Google Gemini AI' terms of service.\n")
	fmt.Printf("  All input data will be used by Google to improve 'Gemini AI'.\n")
	fmt.Printf("  Therefore, do not process any private or confidential data.\n")

	fmt.Printf("\nInput from:\n")
	if progConfig.InputFromTerminal {
		fmt.Printf("  Terminal  : yes\n")
	}
	if progConfig.InputFromFile {
		fmt.Printf("  File      : %v\n", progConfig.InputFile)
	}
	if progConfig.InputFromLocalhost {
		fmt.Printf("  localhost : %v (port)\n", progConfig.InputLocalhostPort)
	}

	fmt.Printf("\nRendering:\n")
	fmt.Printf("  Markdown : %v\n", progConfig.MarkdownPromptResponseFile)
	if progConfig.AnsiRendering {
		fmt.Printf("  Ansi     : %v\n", progConfig.AnsiPromptResponseFile)
	}
	if progConfig.HTMLRendering {
		fmt.Printf("  HTML     : %v\n", progConfig.HTMLPromptResponseFile)
	}

	fmt.Printf("\nHistory:\n")
	if progConfig.MarkdownHistory {
		fmt.Printf("  Markdown : %v\n", progConfig.MarkdownHistoryDirectory)
	}
	if progConfig.AnsiHistory {
		fmt.Printf("  Ansi     : %v\n", progConfig.AnsiHistoryDirectory)
	}
	if progConfig.HTMLHistory {
		fmt.Printf("  HTML     : %v\n", progConfig.HTMLHistoryDirectory)
	}

	fmt.Printf("\nOutput:\n")
	if progConfig.AnsiOutput {
		fmt.Printf("  Terminal : yes\n")
	}
	if progConfig.MarkdownOutput {
		fmt.Printf("  Markdown : execute application\n")
	}
	if progConfig.HTMLOutput {
		fmt.Printf("  HTML     : execute application\n")
	}

	if len(progConfig.MIMETypeReplacements) > 0 {
		fmt.Printf("\nMIME Type Replacements:\n")
		for _, replacement := range progConfig.MIMETypeReplacements {
			fmt.Printf("  %s\n", replacement)
		}
	}
}

/*
initializeProgram performs program initialization tasks. It sets up the program environment, including
creating necessary directories and writing assets for HTML output.
*/
func initializeProgram() {
	var err error

	// create history directories
	if progConfig.MarkdownHistory {
		err = os.Mkdir(progConfig.MarkdownHistoryDirectory, 0750)
		if err != nil && !os.IsExist(err) {
			fmt.Printf("error [%v] at os.Mkdir()\n", err)
			os.Exit(1)
		}
	}
	if progConfig.AnsiHistory {
		err = os.Mkdir(progConfig.AnsiHistoryDirectory, 0750)
		if err != nil && !os.IsExist(err) {
			fmt.Printf("error [%v] at os.Mkdir()\n", err)
			os.Exit(1)
		}
	}
	if progConfig.HTMLHistory {
		err = os.Mkdir(progConfig.HTMLHistoryDirectory, 0750)
		if err != nil && !os.IsExist(err) {
			fmt.Printf("error [%v] at os.Mkdir()\n", err)
			os.Exit(1)
		}

		// 'assets' in history directory (to render HTML files in history directory)
		directory := progConfig.HTMLHistoryDirectory + "/assets"
		if !dirExists(directory) {
			err = os.Mkdir(directory, 0750)
			if err != nil && !os.IsExist(err) {
				fmt.Printf("error [%v] at os.Mkdir()\n", err)
				os.Exit(1)
			}
			writeAssets(progConfig.HTMLHistoryDirectory)
		}
	}
}

/*
// application-specific, static part of "System-Prompt"
const appSystemInstruction = `You are a CLI backend processor for 'gem-pro'.
Your goal: Provide helpful content first, then generate a strict archiving slug.

### FORMATTING RULES
- Provide raw markdown only.
- No conversational filler (e.g., "Sure, I can help with that").
- Do not wrap the entire response in triple backticks (code blocks).
- Ensure LaTeX math blocks ($$) are always placed on their own new lines.

### METADATA & ARCHIVING
- You must generate a metadata line at the very end of your response.
- Create a filename slug based on your generated content.
- Length: 3-6 words.
- Format: Lowercase, ASCII only, hyphens for spaces (kebab-case).
- Language Rules:
  - Keep main language of response if Latin script used.
  - Transliteration: ä->ae, ö->oe, ü->ue, ß->ss, ñ->n, etc.
  - Translation: If content is Non-Latin (Cyrillic, CJK, Arabic, etc.), translate the slug to ENGLISH.
  - Remove all emojis/symbols.

### OUTPUT FORMAT
[Actual Content Here]
...
[End of Content]

METADATA_SLUG: <your-slug-here>`
*/

// application-specific, static part of "System-Prompt"
const appSystemInstruction = `Objective: Provide helpful, accurate content followed by a metadata archiving line.

Constraints:
1. Response Content: Use raw Markdown only.
2. No Conversational Filler: Do not include introductory phrases (e.g., "Here is the information").
3. No Outer Wrappers: Do NOT wrap the entire response in triple backticks (e.g., no ` + "```" + `markdown ... ` + "```" + `). Only use backticks for code examples within the content.
4. Mathematical Notation: 
   - Use inline math ($...$) ONLY for single variables or simple values (e.g., $x$, $50\%$).
   - Use display math ($$...$$) ONLY for complex formulas. 
   - Every display math block MUST be preceded and followed by a completely empty newline.
   - Avoid LaTeX for plain text or simple arithmetic to ensure renderer stability.
5. Archiving: The very last line of the response must be the metadata slug.

Metadata Slug Rules:
- Format: "METADATA_SLUG: kebab-case-slug"
- Content: 3-6 descriptive words.
- Characters: Lowercase ASCII, hyphens only (no symbols/emojis).
- Localization: Transliterate special characters (ä->ae, etc.). Translate non-Latin content to English for the slug.

Output Structure:
<Content>`

/*
generateGeminiModelConfig generates a configuration object for the Gemini AI model. It creates a
genai.GenerateContentConfig object and configures it based on the program settings for interacting
with the Gemini AI model.
*/
func generateGeminiModelConfig(isImageRequest bool, cacheName string, storeNames []string) *genai.GenerateContentConfig {
	generateContentConfig := &genai.GenerateContentConfig{
		// HTTPOptions *HTTPOptions
		// SystemInstruction *Content
		// Temperature *float32
		// TopP *float32
		// TopK *float32
		// CandidateCount int32
		// MaxOutputTokens int32
		// StopSequences []string
		// ResponseLogprobs bool
		// Logprobs *int32
		// PresencePenalty
		// FrequencyPenalty
		// Seed *int32
		// ResponseMIMEType string
		ResponseMIMEType: "text/plain", // always expected: plain text with markdown tags
		// ResponseSchema *Schema
		// ResponseJsonSchema any
		// RoutingConfig *GenerationConfigRoutingConfig
		// ModelSelectionConfig *ModelSelectionConfig
		// SafetySettings []*SafetySetting
		// Tools []*Tool
		// ToolConfig *ToolConfig
		// Labels map[string]string
		// CachedContent string
		// ResponseModalities []string
		// MediaResolution MediaResolution
		// SpeechConfig *SpeechConfig
		// AudioTimestamp bool
		// ThinkingConfig *ThinkingConfig
		// ImageConfig *ImageConfig
	}
	// configure AI model parameters
	if progConfig.GeminiCandidateCount != nil {
		generateContentConfig.CandidateCount = *progConfig.GeminiCandidateCount
	}
	if len(progConfig.GeminiResponseModalities) > 0 {
		generateContentConfig.ResponseModalities = append(generateContentConfig.ResponseModalities, progConfig.GeminiResponseModalities...)
	}
	if progConfig.GeminiTemperature != nil {
		generateContentConfig.Temperature = progConfig.GeminiTemperature
	}
	if progConfig.GeminiTopP != nil {
		generateContentConfig.TopP = progConfig.GeminiTopP
	}
	if progConfig.GeminiTopK != nil {
		generateContentConfig.TopK = progConfig.GeminiTopK
	}
	if progConfig.GeminiMaxOutputTokens != nil && *progConfig.GeminiMaxOutputTokens > 0 {
		generateContentConfig.MaxOutputTokens = *progConfig.GeminiMaxOutputTokens
	}

	// system prompt composition (application part + optional user part)
	finalSystemInstruction = appSystemInstruction
	if progConfig.UserSystemInstruction != "" {
		userInstructionBytes, err := os.ReadFile(progConfig.UserSystemInstruction)
		if err != nil {
			fmt.Printf("error [%v] reading user system instruction file\n", err)
			os.Exit(1)
		}

		// Combine: App Prompt + Header + User Prompt
		if len(userInstructionBytes) > 0 {
			finalSystemInstruction += "\n\nUser Context:\n" + string(userInstructionBytes)
		}
	}
	generateContentConfig.SystemInstruction = genai.NewContentFromText(finalSystemInstruction, "user")

	generateContentConfig.Tools = []*genai.Tool{}
	if progConfig.GeminiGroundingWithCodeExecution {
		generateContentConfig.Tools = append(generateContentConfig.Tools, &genai.Tool{CodeExecution: &genai.ToolCodeExecution{}})
	}
	if progConfig.GeminiGroundingWithGoogleSearch {
		generateContentConfig.Tools = append(generateContentConfig.Tools, &genai.Tool{GoogleSearch: &genai.GoogleSearch{}})
	}
	if progConfig.GeminiGroundingWithURLContext {
		generateContentConfig.Tools = append(generateContentConfig.Tools, &genai.Tool{URLContext: &genai.URLContext{}})
	}
	if progConfig.GeminiGroundigWithGoogleMaps {
		generateContentConfig.Tools = append(generateContentConfig.Tools, &genai.Tool{GoogleMaps: &genai.GoogleMaps{}})
	}
	if len(storeNames) > 0 {
		generateContentConfig.Tools = append(generateContentConfig.Tools, &genai.Tool{
			FileSearch: &genai.FileSearch{
				FileSearchStoreNames: storeNames,
			},
		})
	}

	if cacheName != "" {
		generateContentConfig.CachedContent = cacheName
	}

	var thinkingBudget *int32
	thinkingBudget = progConfig.GeminiMaxThinkingBudget

	var thinkingLevel genai.ThinkingLevel
	switch strings.ToLower(progConfig.GeminiThinkingLevel) {
	case "minimal":
		thinkingLevel = genai.ThinkingLevelMinimal
		thinkingBudget = nil
	case "low":
		thinkingLevel = genai.ThinkingLevelLow
		thinkingBudget = nil
	case "medium":
		thinkingLevel = genai.ThinkingLevelMedium
		thinkingBudget = nil
	case "high":
		thinkingLevel = genai.ThinkingLevelHigh
		thinkingBudget = nil
	}

	generateContentConfig.ThinkingConfig = &genai.ThinkingConfig{
		IncludeThoughts: progConfig.GeminiIncludeThoughts,
		ThinkingBudget:  thinkingBudget,
		ThinkingLevel:   thinkingLevel,
	}

	switch strings.ToLower(progConfig.GeminiInputMediaResolution) {
	case "low":
		generateContentConfig.MediaResolution = genai.MediaResolutionLow
	case "medium":
		generateContentConfig.MediaResolution = genai.MediaResolutionMedium
	case "high":
		generateContentConfig.MediaResolution = genai.MediaResolutionHigh
	}

	/*
		type ImageConfig struct {
			AspectRatio              string
			ImageSize                string
			OutputMIMEType           string
			OutputCompressionQuality *int32
		}
	*/
	if isImageRequest && (progConfig.GeminiImageAspectRatio != "" || progConfig.GeminiImageResolution != "") {
		generateContentConfig.ImageConfig = &genai.ImageConfig{}
		if progConfig.GeminiImageAspectRatio != "" {
			generateContentConfig.ImageConfig.AspectRatio = progConfig.GeminiImageAspectRatio
		}
		if progConfig.GeminiImageResolution != "" {
			generateContentConfig.ImageConfig.ImageSize = progConfig.GeminiImageResolution
		}
	}

	return generateContentConfig
}

/*
printGeminiModelConfig prints relevant parts of the Gemini model configuration to the console. It takes a
Gemini model configuration and prints its key parameters to the terminal, formatted for readability within
a specified width.
*/
func printGeminiModelConfig(geminiModelConfig *genai.GenerateContentConfig, terminalWidth int) {
	fmt.Printf("\nGemini model configuration (excerpt):\n")
	if geminiModelConfig.SystemInstruction != nil {
		if len(geminiModelConfig.SystemInstruction.Parts) > 0 {
			fmt.Printf("  SystemInstruction : %v\n", wrapString(geminiModelConfig.SystemInstruction.Parts[0].Text, terminalWidth, 22))
		}
	}
	if geminiModelConfig.Temperature != nil {
		fmt.Printf("  Temperature       : %v\n", *geminiModelConfig.Temperature)
	}
	if geminiModelConfig.TopP != nil {
		fmt.Printf("  TopP              : %v\n", *geminiModelConfig.TopP)
	}
	if geminiModelConfig.TopK != nil {
		fmt.Printf("  TopK              : %v\n", *geminiModelConfig.TopK)
	}
	fmt.Printf("  CandidateCount    : %v\n", geminiModelConfig.CandidateCount)
	if geminiModelConfig.MaxOutputTokens > 0 {
		fmt.Printf("  MaxOutputTokens   : %v\n", geminiModelConfig.MaxOutputTokens)
	}
	if geminiModelConfig.ResponseMIMEType != "" {
		fmt.Printf("  ResponseMIMEType  : %v\n", geminiModelConfig.ResponseMIMEType)
	}
	if geminiModelConfig.Tools != nil {
		for _, tool := range geminiModelConfig.Tools {
			if tool.GoogleSearch != nil {
				fmt.Printf("  Tool              : GoogleSearch\n")
			}
			if tool.URLContext != nil {
				fmt.Printf("  Tool              : URLContext\n")
			}
			if tool.GoogleMaps != nil {
				fmt.Printf("  Tool              : GoogleMaps\n")
			}
			if tool.CodeExecution != nil {
				fmt.Printf("  Tool              : CodeExecution\n")
			}
			if tool.FileSearch != nil {
				// TODO: formatting (separate lines for each store?)
				fmt.Printf("  Tool              : FileSearchStores: %s\n",
					strings.Join(tool.FileSearch.FileSearchStoreNames, ", "))
			}
		}
	}
	if geminiModelConfig.ResponseModalities != nil {
		fmt.Printf("  ResponseModalities: %v\n", strings.Join(geminiModelConfig.ResponseModalities, ", "))
	}
	if geminiModelConfig.CachedContent != "" {
		fmt.Printf("  CachedContent     : %v\n", geminiModelConfig.CachedContent)
	}
	if progConfig.GeminiMaxThinkingBudget != nil {
		fmt.Printf("  ThinkingBudget    : %d\n", *progConfig.GeminiMaxThinkingBudget)
	}
	if progConfig.GeminiThinkingLevel != "" {
		fmt.Printf("  ThinkingLevel     : %s\n", progConfig.GeminiThinkingLevel)
	}
	if progConfig.GeminiInputMediaResolution != "" {
		fmt.Printf("  MediaResolution   : %s\n", progConfig.GeminiInputMediaResolution)
	}
	if geminiModelConfig.ImageConfig != nil {
		if geminiModelConfig.ImageConfig.AspectRatio != "" {
			fmt.Printf("  ImageAspectRatio  : %s\n", geminiModelConfig.ImageConfig.AspectRatio)
		}
		if geminiModelConfig.ImageConfig.ImageSize != "" {
			fmt.Printf("  ImageSize         : %s\n", geminiModelConfig.ImageConfig.ImageSize)
		}
	}
}

/*
showCompactConfiguration shows a very compact overview of the most important parameters.
*/
func showCompactConfiguration(modelInfo *genai.Model, modelConfig *genai.GenerateContentConfig) {
	fmt.Printf("\n--- %s %s ---------------------------------------------------\n", progName, progVersion)

	// Modell & Limits
	inLimit := fmt.Sprintf("%dk", modelInfo.InputTokenLimit/1024)
	if modelInfo.InputTokenLimit >= 1048576 {
		inLimit = fmt.Sprintf("%dM", modelInfo.InputTokenLimit/1048576)
	}
	outLimit := fmt.Sprintf("%dk", modelInfo.OutputTokenLimit/1024)
	fmt.Printf("Model  : %s (Limits: %s In / %s Out)\n", modelInfo.Name, inLimit, outLimit)

	// Config
	configParts := []string{fmt.Sprintf("%d Candidate(s)", *progConfig.GeminiCandidateCount)}
	if progConfig.GeminiThinkingLevel != "" {
		configParts = append(configParts, "Thinking: "+progConfig.GeminiThinkingLevel)
	}
	if progConfig.GeminiInputMediaResolution != "" {
		configParts = append(configParts, "MediaResolution: "+progConfig.GeminiInputMediaResolution)
	}
	fmt.Printf("Config : %s\n", strings.Join(configParts, ", "))

	// Tools
	var activeTools []string
	if progConfig.GeminiGroundingWithGoogleSearch {
		activeTools = append(activeTools, "GoogleSearch")
	}
	if progConfig.GeminiGroundingWithURLContext {
		activeTools = append(activeTools, "URLContext")
	}
	if progConfig.GeminiGroundingWithCodeExecution {
		activeTools = append(activeTools, "CodeExecution")
	}
	if progConfig.GeminiGroundigWithGoogleMaps {
		activeTools = append(activeTools, "GoogleMaps")
	}
	if len(activeTools) > 0 {
		fmt.Printf("Tools  : %s\n", strings.Join(activeTools, ", "))
	}

	// Context (Files, Cache, RAG)
	okCount := 0
	warnCount := 0
	errorCount := 0
	for _, f := range filesToHandle {
		switch f.State {
		case "ok":
			okCount++
		case "warn":
			warnCount++
		case "error":
			errorCount++
		}
	}

	if len(filesToHandle) > 0 {
		loadedCount := okCount + warnCount
		infoPart := fmt.Sprintf("Files  : %d local files loaded", loadedCount)

		var details []string
		if warnCount > 0 {
			details = append(details, fmt.Sprintf("%d %s", warnCount, pluralize(warnCount, "warning")))
		}
		if errorCount > 0 {
			details = append(details, fmt.Sprintf("%d %s", errorCount, pluralize(errorCount, "error")))
		}

		if len(details) > 0 {
			infoPart += fmt.Sprintf(" (%s)", strings.Join(details, ", "))
		}
		fmt.Printf("%s\n", infoPart)
	}

	if *includeFiles {
		fmt.Printf("Remote : Included files from Google File Store\n")
	}
	if modelConfig.CachedContent != "" {
		fmt.Printf("Cache  : %s\n", modelConfig.CachedContent)
	}
	if len(includeStores) > 0 {
		fmt.Printf("RAG    : %s\n", strings.Join(includeStores, ", "))
	}

	// Mode
	modeStr := "Non-Chat"
	if *chatmode {
		modeStr = "Chat"
	}
	if progConfig.GeminiPureResponse {
		modeStr += ", Pure-Response"
	}
	fmt.Printf("Mode   : %s\n", modeStr)

	// Output
	var outputs []string
	if progConfig.AnsiOutput {
		outputs = append(outputs, "Terminal")
	}
	if progConfig.HTMLOutput {
		outputs = append(outputs, "HTML")
	}
	if progConfig.MarkdownOutput {
		outputs = append(outputs, "Markdown")
	}
	fmt.Printf("Output : %s\n", strings.Join(outputs, ", "))

	fmt.Printf("Quit   : CTRL-C\n")
	fmt.Printf("-----------------------------------------------------------------------\n")
	fmt.Printf("Free Tier: Data is used by Google to improve AI. No confidential data!\n")
	fmt.Printf("Paid Tier: Data is private and not used for training (GCP terms apply).\n")
	fmt.Printf("-----------------------------------------------------------------------\n\n")
}
