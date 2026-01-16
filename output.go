package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"google.golang.org/genai"
)

/*
printPromptResponseToTerminal prints the content of the ANSI prompt/response file to the standard output (terminal).
It reads the content from the ANSI formatted prompt / response file and writes it directly to the standard output,
displaying colored text in the terminal.
*/
func printPromptResponseToTerminal() {
	data, err := os.ReadFile(progConfig.AnsiPromptResponseFile)
	if err != nil {
		fmt.Printf("error [%v] at os.ReadFile()\n", err)
		return
	}
	_, _ = os.Stdout.Write(data)
}

/*
processPrompt processes the user prompt and prepares it for different output formats (Markdown, ANSI, HTML).
It takes a user prompt, formats it into Markdown, ANSI, and HTML, including system instructions and referenced
files, and saves these formats to respective files.
*/
func processPrompt(prompt string, chatmode bool, chatNumber int) {
	// If pure response is requested, do not write prompt to output files.
	// But ensure files are empty/truncated so they don't contain old data.
	if progConfig.GeminiPureResponse {
		_ = os.WriteFile(progConfig.MarkdownPromptResponseFile, []byte(""), 0600)
		_ = os.WriteFile(progConfig.AnsiPromptResponseFile, []byte(""), 0600)
		_ = os.WriteFile(progConfig.HTMLPromptResponseFile, []byte(""), 0600)
		return
	}

	var promptString strings.Builder

	// text part of prompt (also included in contents)
	promptString.WriteString("***\n")
	if chatmode {
		if chatNumber == 1 {
			promptString.WriteString("**Prompt to Gemini (initial chat #1):**\n")
		} else {
			promptString.WriteString(fmt.Sprintf("**Prompt to Gemini (refinement chat #%d):**\n", chatNumber))
		}
	} else {
		promptString.WriteString("**Prompt to Gemini:**\n")
	}
	promptString.WriteString("\n```plaintext\n")
	promptString.WriteString(prompt)
	promptString.WriteString("\n```\n")
	promptString.WriteString("\n***\n")

	// system instructions part of prompt (not included in contents, but important)
	if progConfig.GeminiSystemInstruction != "" {
		promptString.WriteString("**System Instruction to Gemini:**\n")
		promptString.WriteString("\n```plaintext\n")
		promptString.WriteString(progConfig.GeminiSystemInstruction)
		promptString.WriteString("\n```\n")
		promptString.WriteString("\n***\n")
	}

	if (chatmode && chatNumber == 1) || !chatmode {
		if len(filesToHandle) > 0 {
			promptString.WriteString("**Data referenced by the Prompt (from commandline):**\n")
			promptString.WriteString("\n```plaintext\n")
			for _, fileToUpload := range filesToHandle {
				if fileToUpload.State != "error" {
					// add replacement MIME type (e.g. 'text/x-perl -> text/plain')
					mimeType := fileToUpload.MimeType
					if ReplacementMIMETypeMap != nil {
						replacement, ok := ReplacementMIMETypeMap[fileToUpload.MimeType]
						if ok {
							mimeType += fmt.Sprintf(" -> %s", replacement)
						}
					}
					promptString.WriteString(fmt.Sprintf("%-5s %s (%s, %s, %s)\n",
						fileToUpload.State, fileToUpload.Filepath, fileToUpload.LastUpdate, fileToUpload.FileSize, mimeType))
				} else {
					promptString.WriteString(fmt.Sprintf("%-5s %s %s\n",
						fileToUpload.State, fileToUpload.Filepath, fileToUpload.ErrorMessage))
				}
			}
			promptString.WriteString("```\n")
			promptString.WriteString("\n***\n")
		}

		if *includeFiles {
			promptString.WriteString("**Data referenced by the Prompt (from Google file store):**\n")
			promptString.WriteString("\n```plaintext\n")
			promptString.WriteString(listFilesUploadedToGemini(""))
			promptString.WriteString("```\n")
			promptString.WriteString("\n***\n")
		}

		if *includeCache {
			promptString.WriteString("**Data referenced by the Prompt (from AI model cache):**\n")
			promptString.WriteString("\n```plaintext\n")
			_, cacheDetails := listAIModelSpecificCache("")
			promptString.WriteString(cacheDetails)
			promptString.WriteString("```\n")
			promptString.WriteString("\n***\n")
		}

		if len(includeStores) > 0 {
			promptString.WriteString("**Data referenced by the Prompt (from FileSearchStores):**\n")
			promptString.WriteString("\n```plaintext\n")
			for _, storeID := range includeStores {
				promptString.WriteString(fmt.Sprintf("Included Store: %s\n", storeID))
			}
			promptString.WriteString("```\n")
			promptString.WriteString("\n***\n")
		}
	}

	// write prompt to current markdown request/response file
	err := os.WriteFile(progConfig.MarkdownPromptResponseFile, []byte(promptString.String()), 0600)
	if err != nil {
		fmt.Printf("error [%v] at os.WriteFile()\n", err)
		return
	}

	// render prompt as ansi
	ansiData := promptString.String()
	if progConfig.AnsiRendering {
		ansiData = renderMarkdown2Ansi(promptString.String())
	}

	// write prompt to current ansi request/response file
	err = os.WriteFile(progConfig.AnsiPromptResponseFile, []byte(ansiData), 0600)
	if err != nil {
		fmt.Printf("error [%v] at os.WriteFile()\n", err)
		return
	}

	// render prompt as html
	htmlData := promptString.String()
	if progConfig.HTMLRendering {
		htmlData = renderMarkdown2HTML(promptString.String())
	}

	// write prompt to current html request/response file
	err = os.WriteFile(progConfig.HTMLPromptResponseFile, []byte(htmlData), 0600)
	if err != nil {
		fmt.Printf("error [%v] at os.WriteFile()\n", err)
		return
	}
}

/*
getCandidateText extracts the text content from a candidate.
If includeThoughts is true, it wraps thoughts in the specific HTML/Markdown block used by this application.
If includeThoughts is false, thoughts are skipped.
*/
func getCandidateText(candidate *genai.Candidate, includeThoughts bool) string {
	if candidate.Content == nil {
		return "No content available in this candidate.\n"
	}

	var sb strings.Builder
	var aggregatedThoughts strings.Builder
	var regularContent strings.Builder

	for _, part := range candidate.Content.Parts {
		if part.Thought {
			if includeThoughts && part.Text != "" {
				aggregatedThoughts.WriteString(strings.TrimSpace(part.Text) + "\n\n")
			}
			continue
		}

		// regular content (anything that isn't a 'thought')
		if part.VideoMetadata != nil {
			regularContent.WriteString("Metadata for a given video.\n")
		}
		if part.CodeExecutionResult != nil {
			regularContent.WriteString("\nCode Execution Result:\n")
			regularContent.WriteString("\n```plaintext\n")
			if part.CodeExecutionResult.Outcome != genai.OutcomeOK {
				regularContent.WriteString(fmt.Sprintf("%s\n\n", part.CodeExecutionResult.Outcome))
			}
			regularContent.WriteString(strings.TrimSuffix(part.CodeExecutionResult.Output, "\n"))
			regularContent.WriteString("\n```\n")
		}
		if part.ExecutableCode != nil {
			regularContent.WriteString(fmt.Sprintf("\nExecutable %s Code:\n", part.ExecutableCode.Language))
			regularContent.WriteString(fmt.Sprintf("\n```%s\n", part.ExecutableCode.Language))
			regularContent.WriteString(strings.TrimSuffix(part.ExecutableCode.Code, "\n"))
			regularContent.WriteString("\n```\n")
		}
		if part.FileData != nil {
			regularContent.WriteString(fmt.Sprintf("File Data: URI=%s, MIME=%s\n", part.FileData.FileURI, part.FileData.MIMEType))
		}
		if part.FunctionCall != nil {
			regularContent.WriteString("A predicted [FunctionCall] returned from the model.\n")
		}
		if part.FunctionResponse != nil {
			regularContent.WriteString("The result output of a [FunctionCall].\n")
		}
		if part.InlineData != nil {
			regularContent.WriteString(fmt.Sprintf("Inline data (%.1f KiB, %s) : ", float64(len(part.InlineData.Data))/1024.0, part.InlineData.MIMEType))
			pathname, filename, err := writeDataToFile(part.InlineData.Data, part.InlineData.MIMEType, finishProcessing)
			if err != nil {
				regularContent.WriteString(fmt.Sprintf("error [%v] writing data to file\n", err))
			} else {
				u := url.URL{
					Scheme: "file",
					Path:   pathname,
				}
				encodedURL := u.String()
				regularContent.WriteString(fmt.Sprintf("\n![%s](%s)\n\n", filename, encodedURL))
			}
		}
		if part.Text != "" { // ensure that part.Text is not from Thought
			regularContent.WriteString(removeSpacesBetweenNewlineAndCodeblock(part.Text))
			regularContent.WriteString("\n")
		}
	}

	// append thoughts block if requested and available
	if includeThoughts && aggregatedThoughts.Len() > 0 {
		sb.WriteString("<!-- AI_THOUGHT_SUMMARY_START -->")
		sb.WriteString("<!-- AI_THOUGHT_SUMMARY_END -->\n")
		sb.WriteString("<!-- AI_THOUGHT_CONTENT_START -->\n")
		sb.WriteString(strings.TrimSpace(aggregatedThoughts.String()) + "\n")
		sb.WriteString("<!-- AI_THOUGHT_CONTENT_END -->\n\n")
	}

	// append regular content
	sb.WriteString(regularContent.String())
	sb.WriteString("\n")

	return sb.String()
}

/*
processPureResponse processes the Gemini AI model's response and formats it for output.
It extracts content from candidates without adding boilerplate metadata.
*/
func processPureResponse(resp *genai.GenerateContentResponse) {
	var responseString strings.Builder

	// print response candidate(s)
	for _, candidate := range resp.Candidates {
		// Get text content, explicitly excluding thoughts
		responseString.WriteString(getCandidateText(candidate, false))

		// show why the model stopped generating tokens (content)
		if candidate.FinishReason != genai.FinishReasonStop {
			responseString.WriteString("\n***\n")
			responseString.WriteString(fmt.Sprintf("Model stopped generating tokens (content) with reason [%s].\n", candidate.FinishReason))
		}
	}

	// append response string to request/response files
	appendResponseString(responseString)
}

/*
processResponse processes the Gemini AI model's response and formats it for output.
It includes headers, thoughts (if configured), citations, grounding, and metadata.
*/
func processResponse(resp *genai.GenerateContentResponse) {
	var responseString strings.Builder

	// print response candidate(s)
	for i, candidate := range resp.Candidates {
		if len(resp.Candidates) > 1 {
			responseString.WriteString(fmt.Sprintf("**Response from Gemini (Candidate #%d):**\n\n", (i + 1)))
		} else {
			responseString.WriteString("**Response from Gemini:**\n\n")
		}

		// Get text content, including thoughts based on config (Thoughts are part of the 'text' logic in getCandidateText)
		// Note: progConfig.GeminiIncludeThoughts ensures we receive them from API, passing 'true' here formats them.
		responseString.WriteString(getCandidateText(candidate, true))

		// build list of text citation source URIs
		citationURIs := []string{}
		if candidate.CitationMetadata != nil {
			for _, citation := range candidate.CitationMetadata.Citations {
				if citation.URI != "" {
					citationURIs = append(citationURIs, (fmt.Sprintf("%v", citation.URI)))
				}
			}
		}

		// show text citation source URIs
		if len(citationURIs) > 0 {
			responseString.WriteString("\n***\n")
			responseString.WriteString(fmt.Sprintf("Text Citation %s:\n\n", pluralize(len(citationURIs), "Source")))
			for _, citationURI := range citationURIs {
				responseString.WriteString(fmt.Sprintf("* [%s](%s)\n", citationURI, citationURI))
			}
		}

		// build list of code citation licenses
		citationLicenses := []string{}
		if candidate.CitationMetadata != nil {
			for _, citation := range candidate.CitationMetadata.Citations {
				if citation.License != "" {
					citationLicenses = append(citationLicenses, citation.License)
				}
			}
		}

		// show code citation licenses (needs revision, output never seen)
		if len(citationLicenses) > 0 {
			responseString.WriteString("\n***\n")
			responseString.WriteString(fmt.Sprintf("Code Citation %s:\n\n", pluralize(len(citationLicenses), "License")))
			for _, citationSourceLicense := range citationLicenses {
				responseString.WriteString(fmt.Sprintf("* %s\n", citationSourceLicense))
			}
		}

		// show why the model stopped generating tokens (content) (needs revision, output never seen)
		if candidate.FinishReason != genai.FinishReasonStop {
			responseString.WriteString("\n***\n")
			responseString.WriteString(fmt.Sprintf("Model stopped generating tokens (content) with reason [%s].\n", candidate.FinishReason))
		}

		// show grounding metadata
		if candidate.GroundingMetadata != nil {
			// grounding: show list of used web resources (search sources)
			if candidate.GroundingMetadata.GroundingChunks != nil {
				responseString.WriteString("\n***\n")
				responseString.WriteString("**Online Search Sources Used:**\n\n")
				// numbered list because response can contain references (e.g. [2] or [1,3,15])
				for k, groundingChunk := range candidate.GroundingMetadata.GroundingChunks {
					switch {
					case groundingChunk.Web != nil:
						responseString.WriteString(fmt.Sprintf("%d. [%s](%s)\n", k+1, groundingChunk.Web.Title, groundingChunk.Web.URI))
					case groundingChunk.Maps != nil:
						responseString.WriteString(fmt.Sprintf("%d. [%s](%s)\n", k+1, groundingChunk.Maps.Title, groundingChunk.Maps.URI))
					case groundingChunk.RetrievedContext != nil:
						responseString.WriteString(fmt.Sprintf("%d. [%s](%s)\n", k+1, groundingChunk.RetrievedContext.Title, groundingChunk.RetrievedContext.URI))
					}
				}
			}
			// grounding: show list of recommended web search queries (google search suggestions)
			if candidate.GroundingMetadata.WebSearchQueries != nil {
				responseString.WriteString("\n***\n")
				responseString.WriteString("**Google Search Suggestions:**\n\n")
				for _, webSearchQuery := range candidate.GroundingMetadata.WebSearchQueries {
					responseString.WriteString(fmt.Sprintf("* [%s](https://www.google.com/search?q=%s)\n", webSearchQuery, url.QueryEscape(webSearchQuery)))
				}
			}
		}
		responseString.WriteString("\n***\n")
	}

	temperatureInfo := "Temperature: default"
	if progConfig.GeminiTemperature != nil {
		temperatureInfo = fmt.Sprintf("Temperature: %.2f", *progConfig.GeminiTemperature)
	}
	toppInfo := "TopP: default"
	if progConfig.GeminiTopP != nil {
		toppInfo = fmt.Sprintf("TopP: %.2f", *progConfig.GeminiTopP)
	}

	// print response metadata
	responseString.WriteString("```plaintext\n")
	responseString.WriteString(fmt.Sprintf("AI model   : %v (%s, %s)\n", resp.ModelVersion, temperatureInfo, toppInfo))

	var activeTools []string
	if progConfig.GeminiGroundingWithGoogleSearch {
		activeTools = append(activeTools, "Google Search")
	}
	if progConfig.GeminiGroundingWithURLContext {
		activeTools = append(activeTools, "URLContext")
	}
	if progConfig.GeminiGroundingWithCodeExecution {
		activeTools = append(activeTools, "Code Execution")
	}
	if progConfig.GeminiGroundigWithGoogleMaps {
		activeTools = append(activeTools, "Google Maps")
	}
	if len(includeStores) > 0 {
		activeTools = append(activeTools, "FileSearchStores")
	}

	if len(activeTools) > 0 {
		responseString.WriteString(fmt.Sprintf("Tools      : %s\n", strings.Join(activeTools, ", ")))
	}

	responseString.WriteString(fmt.Sprintf("Generated  : %v\n", finishProcessing.Format(time.RFC850)))

	duration := finishProcessing.Sub(startProcessing)
	responseString.WriteString(fmt.Sprintf("Processing : %.1f secs for %d %s\n", duration.Seconds(),
		len(resp.Candidates), pluralize(len(resp.Candidates), "candidate")))

	/*
	   Cost Calculation Logic (Corrected for SDK v0.8.0+ / Gemini 1.5+ / Gemini 3):
	   Total Tokens = Prompt + Tools + Candidates + Thoughts
	   1. Input Total: PromptTokenCount + ToolUsePromptTokenCount
	   2. Net Prompt : PromptTokenCount - CachedContentTokenCount
	   3. Tools      : ToolUsePromptTokenCount
	   4. Output     : CandidatesTokenCount + ThoughtsTokenCount
	*/
	if resp.UsageMetadata != nil {
		u := resp.UsageMetadata

		// Output Header: Total Tokens
		responseString.WriteString(fmt.Sprintf("Tokens     : %d (Total)\n", u.TotalTokenCount))

		// 1. Input Group
		// Calculate absolute Input Total (Text + Tools)
		totalInputCount := u.PromptTokenCount + u.ToolUsePromptTokenCount

		// Calculate "Net" Prompt (New/Uncached Text)
		// Assumption: Cached content is a subset of the standard PromptTokenCount.
		netPromptCount := u.PromptTokenCount - u.CachedContentTokenCount
		if netPromptCount < 0 {
			netPromptCount = 0 // Safety fallback
		}

		inputDetails := []string{fmt.Sprintf("Prompt: %d", netPromptCount)}

		// Only show Tools/Cached if they are actually used
		if u.ToolUsePromptTokenCount > 0 {
			inputDetails = append(inputDetails, fmt.Sprintf("Tools: %d", u.ToolUsePromptTokenCount))
		}
		if u.CachedContentTokenCount > 0 {
			inputDetails = append(inputDetails, fmt.Sprintf("Cached: %d", u.CachedContentTokenCount))
		}

		// Display Total Input and breakdown
		responseString.WriteString(fmt.Sprintf("  Input    : %d (%s)\n",
			totalInputCount, strings.Join(inputDetails, ", ")))

		// 2. Output Group
		// We sum them up to show the real total output volume.
		totalOutputCount := u.CandidatesTokenCount + u.ThoughtsTokenCount

		outputDetails := []string{fmt.Sprintf("Candidates: %d", u.CandidatesTokenCount)}

		if u.ThoughtsTokenCount > 0 {
			outputDetails = append(outputDetails, fmt.Sprintf("Thoughts: %d", u.ThoughtsTokenCount))
		}

		responseString.WriteString(fmt.Sprintf("  Output   : %d (%s)\n",
			totalOutputCount, strings.Join(outputDetails, ", ")))
	}

	if resp.PromptFeedback != nil {
		responseString.WriteString(fmt.Sprintf("Blocked    : %v\n", resp.PromptFeedback.BlockReasonMessage))
	}

	responseString.WriteString("```\n")
	responseString.WriteString("\n***\n")

	// append response string to request/response files
	appendResponseString(responseString)
}

/*
processError processes errors received from the Gemini AI model. It handles error responses from the Gemini AI
model, formats the error message in Markdown, and prepares it for output, including metadata about the error.
*/
func processError(err error) {
	var responseString strings.Builder

	// handle error response
	responseString.WriteString("**Error Response from Gemini:**\n\n")
	responseString.WriteString("```\n")
	responseString.WriteString(err.Error())
	responseString.WriteString("\n")

	responseString.WriteString("```\n")
	responseString.WriteString("\n***\n")

	// print response metadata
	responseString.WriteString("```plaintext\n")
	if err == nil {
		responseString.WriteString(fmt.Sprintf("AI model   : %v\n", progConfig.GeminiAiModel))
	}
	responseString.WriteString(fmt.Sprintf("Generated  : %v\n", finishProcessing.Format(time.RFC850)))

	duration := finishProcessing.Sub(startProcessing)
	responseString.WriteString(fmt.Sprintf("Processing : %.1f secs resulting in error\n", duration.Seconds()))

	responseString.WriteString("```\n")
	responseString.WriteString("\n***\n")

	// append response string to request/response files
	appendResponseString(responseString)
}

/*
appendResponseString appends a given response string (which can be a successful response or an error message)
to the current request / response files in Markdown, ANSI, and HTML formats.
*/
func appendResponseString(responseString strings.Builder) {
	originalMarkdownWithHTMLComments := responseString.String()

	// cleanup Markdown
	cleanedMarkdown := cleanMarkdownIndentation(originalMarkdownWithHTMLComments)

	// 1. prepare Markdown for direct file saving (and ANSI rendering)
	// replace HTML comment tags with pure Markdown equivalents
	markdownForFileAndAnsi := cleanedMarkdown
	markdownForFileAndAnsi = strings.ReplaceAll(markdownForFileAndAnsi, "<!-- AI_THOUGHT_SUMMARY_START -->", "**Thoughts - Considerations for answering the prompt:**\n\n")
	markdownForFileAndAnsi = strings.ReplaceAll(markdownForFileAndAnsi, "<!-- AI_THOUGHT_SUMMARY_END -->", "")
	markdownForFileAndAnsi = strings.ReplaceAll(markdownForFileAndAnsi, "<!-- AI_THOUGHT_CONTENT_START -->", "")
	markdownForFileAndAnsi = strings.ReplaceAll(markdownForFileAndAnsi, "<!-- AI_THOUGHT_CONTENT_END -->", "")

	// append response string to current markdown request/response file
	currentFileMarkdown, err := os.OpenFile(progConfig.MarkdownPromptResponseFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("error [%v] at os.OpenFile() for Markdown\n", err)
		return
	}
	defer func() { _ = currentFileMarkdown.Close() }()
	_, err = fmt.Fprint(currentFileMarkdown, markdownForFileAndAnsi)
	if err != nil {
		fmt.Printf("error [%v] writing to Markdown file\n", err)
	}

	// 2. render markdown response as ansi
	ansiData := markdownForFileAndAnsi // use the cleaned version
	if progConfig.AnsiRendering {
		ansiData = renderMarkdown2Ansi(markdownForFileAndAnsi) // pass the cleaned version
	}

	// append response string to current ansi request/response file
	currentFileAnsi, err := os.OpenFile(progConfig.AnsiPromptResponseFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("error [%v] at os.OpenFile() for ANSI\n", err)
	} else {
		defer func() { _ = currentFileAnsi.Close() }()
		_, err = fmt.Fprint(currentFileAnsi, ansiData)
		if err != nil {
			fmt.Printf("error [%v] writing to ANSI file\n", err)
		}
	}

	// 3. render markdown response as html (using original string with comments)
	htmlData := originalMarkdownWithHTMLComments // use the original string with HTML comments
	if progConfig.HTMLRendering {
		// renderMarkdown2HTML will convert the comments to <details>
		htmlData = renderMarkdown2HTML(originalMarkdownWithHTMLComments)
	}

	// append response string to current html request/response file
	currentFileHTML, err := os.OpenFile(progConfig.HTMLPromptResponseFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("error [%v] at os.OpenFile() for HTML\n", err)
	} else {
		defer func() { _ = currentFileHTML.Close() }()
		_, err = fmt.Fprint(currentFileHTML, htmlData)
		if err != nil {
			fmt.Printf("error [%v] writing to HTML file\n", err)
		}
	}
}
