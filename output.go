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
	if chatmode {
		if chatNumber == 1 {
			promptString.WriteString("**Prompt to Gemini (initial chat #1):**\n\n")
		} else {
			fmt.Fprintf(&promptString, "**Prompt to Gemini (refinement chat #%d):**\n\n", chatNumber)
		}
	} else {
		promptString.WriteString("**Prompt to Gemini:**\n\n")
	}
	promptString.WriteString("<!-- PROMPT_USER_START -->\n")
	promptString.WriteString("```plaintext\n")
	promptString.WriteString(prompt)
	promptString.WriteString("\n```\n")
	promptString.WriteString("<!-- PROMPT_USER_END -->\n")

	// system instructions part of prompt (not included in contents, but important)
	if finalSystemInstruction != "" {
		promptString.WriteString("\n<!-- PROMPT_SYSTEM_START -->\n")
		promptString.WriteString("```plaintext\n")
		promptString.WriteString(finalSystemInstruction)
		promptString.WriteString("\n```\n")
		promptString.WriteString("<!-- PROMPT_SYSTEM_END -->\n")
	}

	if (chatmode && chatNumber == 1) || !chatmode {
		hasFiles := len(filesToHandle) > 0 || *includeFiles || *includeCache || len(includeStores) > 0
		if hasFiles {
			promptString.WriteString("\n<!-- PROMPT_RESOURCES_START -->\n")

			// 1. local files as Markdown table
			if len(filesToHandle) > 0 {
				promptString.WriteString("**Local files from commandline**\n\n")
				promptString.WriteString("| State | Path | Size | MIME | Modified |\n")
				promptString.WriteString("| :--- | :--- | :--- | :--- | :--- |\n")
				for _, fileToUpload := range filesToHandle {
					if fileToUpload.State != "error" {
						mimeType := fileToUpload.MimeType
						if ReplacementMIMETypeMap != nil {
							replacement, ok := ReplacementMIMETypeMap[fileToUpload.MimeType]
							if ok {
								mimeType += fmt.Sprintf(" → %s", replacement)
							}
						}
						fmt.Fprintf(&promptString, "| %s | %s | %s | %s | %s |\n",
							fileToUpload.State, fileToUpload.Filepath, fileToUpload.FileSize, mimeType, fileToUpload.LastUpdate)
					} else {
						fmt.Fprintf(&promptString, "| <span style=\"color:red\">%s</span> | %s | - | - | %s |\n",
							fileToUpload.State, fileToUpload.Filepath, fileToUpload.ErrorMessage)
					}
				}
				promptString.WriteString("\n")
			}

			// 2. Google File Store as list
			if *includeFiles {
				promptString.WriteString("#### Google File Store (remote files)\n\n")
				promptString.WriteString("```plaintext\n")
				promptString.WriteString(listFilesUploadedToGemini(""))
				promptString.WriteString("```\n\n")
			}

			// 3. AI Model Cache Details as list
			if *includeCache {
				promptString.WriteString("### AI model cache details\n\n")
				promptString.WriteString("```plaintext\n")
				_, cacheDetails := listAIModelSpecificCache("")
				promptString.WriteString(cacheDetails)
				promptString.WriteString("```\n\n")
			}

			// 4. FileSearchStores (RAG) as list
			if len(includeStores) > 0 {
				promptString.WriteString("### FileSearchStores (RAG knowledge database)\n\n")
				for _, storeID := range includeStores {
					fmt.Fprintf(&promptString, "* Active FileSearchStore: `%s`\n", storeID)
				}
				promptString.WriteString("\n")
			}

			promptString.WriteString("<!-- PROMPT_RESOURCES_END -->\n")
		}
	}
	promptString.WriteString("\n***\n")

	rawPrompt := promptString.String()

	// 1. prepare Markdown for direct file saving (and ANSI rendering)
	markdownForFileAndAnsi := rawPrompt
	markdownForFileAndAnsi = strings.ReplaceAll(markdownForFileAndAnsi, "<!-- PROMPT_USER_START -->", "**User Prompt:**\n")
	markdownForFileAndAnsi = strings.ReplaceAll(markdownForFileAndAnsi, "<!-- PROMPT_USER_END -->", "")

	markdownForFileAndAnsi = strings.ReplaceAll(markdownForFileAndAnsi, "<!-- PROMPT_SYSTEM_START -->", "**System Prompt:**\n")
	markdownForFileAndAnsi = strings.ReplaceAll(markdownForFileAndAnsi, "<!-- PROMPT_SYSTEM_END -->", "")

	markdownForFileAndAnsi = strings.ReplaceAll(markdownForFileAndAnsi, "<!-- PROMPT_RESOURCES_START -->", "**Resources:**\n")
	markdownForFileAndAnsi = strings.ReplaceAll(markdownForFileAndAnsi, "<!-- PROMPT_RESOURCES_END -->", "")

	// write prompt to current markdown request/response file
	err := os.WriteFile(progConfig.MarkdownPromptResponseFile, []byte(markdownForFileAndAnsi), 0600)
	if err != nil {
		fmt.Printf("error [%v] at os.WriteFile()\n", err)
		return
	}

	// render prompt as ansi
	ansiData := markdownForFileAndAnsi
	if progConfig.AnsiRendering {
		ansiData = renderMarkdown2Ansi(markdownForFileAndAnsi)
	}

	// write prompt to current ansi request/response file
	err = os.WriteFile(progConfig.AnsiPromptResponseFile, []byte(ansiData), 0600)
	if err != nil {
		fmt.Printf("error [%v] at os.WriteFile()\n", err)
		return
	}

	// render prompt as html
	htmlData := rawPrompt
	if progConfig.HTMLRendering {
		htmlData = renderMarkdown2HTML(rawPrompt)
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
It returns two strings: the thoughts (if available) and the regular content.
*/
func getCandidateText(candidate *genai.Candidate) (string, string) {
	if candidate.Content == nil {
		return "", "No content available in this candidate.\n"
	}

	var aggregatedThoughts strings.Builder
	var regularContent strings.Builder

	for _, part := range candidate.Content.Parts {
		if part.Thought {
			if part.Text != "" {
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
				fmt.Fprintf(&regularContent, "%s\n\n", part.CodeExecutionResult.Outcome)
			}
			regularContent.WriteString(strings.TrimSuffix(part.CodeExecutionResult.Output, "\n"))
			regularContent.WriteString("\n```\n")
		}
		if part.ExecutableCode != nil {
			fmt.Fprintf(&regularContent, "\nExecutable %s Code:\n", part.ExecutableCode.Language)
			fmt.Fprintf(&regularContent, "\n```%s\n", part.ExecutableCode.Language)
			regularContent.WriteString(strings.TrimSuffix(part.ExecutableCode.Code, "\n"))
			regularContent.WriteString("\n```\n")
		}
		if part.FileData != nil {
			fmt.Fprintf(&regularContent, "File Data: URI=%s, MIME=%s\n", part.FileData.FileURI, part.FileData.MIMEType)
		}
		if part.FunctionCall != nil {
			regularContent.WriteString("A predicted [FunctionCall] returned from the model.\n")
		}
		if part.FunctionResponse != nil {
			regularContent.WriteString("The result output of a [FunctionCall].\n")
		}
		if part.InlineData != nil {
			fmt.Fprintf(&regularContent, "Inline data (%.1f KiB, %s) : ", float64(len(part.InlineData.Data))/1024.0, part.InlineData.MIMEType)
			pathname, filename, err := writeDataToFile(part.InlineData.Data, part.InlineData.MIMEType, finishProcessing)
			if err != nil {
				fmt.Fprintf(&regularContent, "error [%v] writing data to file\n", err)
			} else {
				u := url.URL{
					Scheme: "file",
					Path:   pathname,
				}
				encodedURL := u.String()
				fmt.Fprintf(&regularContent, "\n![%s](%s)\n\n", filename, encodedURL)
			}
		}
		if part.Text != "" { // ensure that part.Text is not from Thought
			regularContent.WriteString(removeSpacesBetweenNewlineAndCodeblock(part.Text))
			regularContent.WriteString("\n")
		}
	}

	return strings.TrimSpace(aggregatedThoughts.String()), regularContent.String()
}

/*
processPureResponse processes the Gemini AI model's response and formats it for output.
It extracts content from candidates without adding boilerplate metadata.
*/
func processPureResponse(resp *genai.GenerateContentResponse) {
	var responseString strings.Builder

	responseString.WriteString("\n")

	// print response candidate(s)
	for _, candidate := range resp.Candidates {
		// Get text content, explicitly excluding thoughts
		_, content := getCandidateText(candidate)
		responseString.WriteString(content)

		// show why the model stopped generating tokens (content)
		if candidate.FinishReason != genai.FinishReasonStop {
			responseString.WriteString("\n***\n")
			fmt.Fprintf(&responseString, "Model stopped generating tokens (content) with reason [%s].\n", candidate.FinishReason)
		}
	}

	responseString.WriteString("\n")

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
			fmt.Fprintf(&responseString, "**Response from Gemini (Candidate #%d):**\n\n", (i + 1))
		} else {
			responseString.WriteString("**Response from Gemini:**\n\n")
		}

		thoughts, content := getCandidateText(candidate)

		if thoughts != "" && progConfig.GeminiIncludeThoughts {
			responseString.WriteString("<!-- THOUGHTS_START -->\n")
			responseString.WriteString(thoughts + "\n")
			responseString.WriteString("<!-- THOUGHTS_END -->\n")
		}

		responseString.WriteString("\n")

		responseString.WriteString("<!-- RESPONSE_START -->\n")

		responseString.WriteString(content)

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
			fmt.Fprintf(&responseString, "Text Citation %s:\n\n", pluralize(len(citationURIs), "Source"))
			for _, citationURI := range citationURIs {
				fmt.Fprintf(&responseString, "* [%s](%s)\n", citationURI, citationURI)
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

		// show code citation licenses
		if len(citationLicenses) > 0 {
			responseString.WriteString("\n***\n")
			fmt.Fprintf(&responseString, "Code Citation %s:\n\n", pluralize(len(citationLicenses), "License"))
			for _, citationSourceLicense := range citationLicenses {
				fmt.Fprintf(&responseString, "* %s\n", citationSourceLicense)
			}
		}

		// show why the model stopped generating tokens (content)
		if candidate.FinishReason != genai.FinishReasonStop {
			responseString.WriteString("\n***\n")
			fmt.Fprintf(&responseString, "Model stopped generating tokens (content) with reason [%s].\n", candidate.FinishReason)
		}

		// show grounding metadata
		if candidate.GroundingMetadata != nil {
			// grounding: show list of used web resources (search sources)
			if candidate.GroundingMetadata.GroundingChunks != nil {
				responseString.WriteString("\n***\n")
				responseString.WriteString("**Online Search Sources Used:**\n\n")
				for k, groundingChunk := range candidate.GroundingMetadata.GroundingChunks {
					switch {
					case groundingChunk.Web != nil:
						fmt.Fprintf(&responseString, "%d. [%s](%s)\n", k+1, groundingChunk.Web.Title, groundingChunk.Web.URI)
					case groundingChunk.Maps != nil:
						fmt.Fprintf(&responseString, "%d. [%s](%s)\n", k+1, groundingChunk.Maps.Title, groundingChunk.Maps.URI)
					case groundingChunk.RetrievedContext != nil:
						fmt.Fprintf(&responseString, "%d. [%s](%s)\n", k+1, groundingChunk.RetrievedContext.Title, groundingChunk.RetrievedContext.URI)
					}
				}
			}
			// grounding: show list of recommended web search queries (google search suggestions)
			if candidate.GroundingMetadata.WebSearchQueries != nil {
				responseString.WriteString("\n***\n")
				responseString.WriteString("**Google Search Suggestions:**\n\n")
				for _, webSearchQuery := range candidate.GroundingMetadata.WebSearchQueries {
					fmt.Fprintf(&responseString, "* [%s](https://www.google.com/search?q=%s)\n", webSearchQuery, url.QueryEscape(webSearchQuery))
				}
			}
		}

		responseString.WriteString("<!-- RESPONSE_END -->\n")

		if i < len(resp.Candidates)-1 {
			responseString.WriteString("\n***\n")
		}
	}

	var modelParams []string
	if progConfig.GeminiTemperature != nil {
		modelParams = append(modelParams, fmt.Sprintf("Temperature: %.2f", *progConfig.GeminiTemperature))
	}
	if progConfig.GeminiTopP != nil {
		modelParams = append(modelParams, fmt.Sprintf("TopP: %.2f", *progConfig.GeminiTopP))
	}
	if progConfig.GeminiTopK != nil {
		modelParams = append(modelParams, fmt.Sprintf("TopK: %.2f", *progConfig.GeminiTopK))
	}
	if progConfig.GeminiThinkingLevel != "" {
		modelParams = append(modelParams, fmt.Sprintf("ThinkingLevel: %s", progConfig.GeminiThinkingLevel))
	}

	serviceTierInfo := "default"
	if progConfig.GeminiServiceTier != "" {
		serviceTierInfo = progConfig.GeminiServiceTier
	}
	modelParams = append(modelParams, fmt.Sprintf("ServiceTier: %s", serviceTierInfo))

	paramsStr := ""
	if len(modelParams) > 0 {
		paramsStr = " (" + strings.Join(modelParams, ", ") + ")"
	}

	// print response metadata
	responseString.WriteString("<!-- STATS_START -->\n")
	responseString.WriteString("```plaintext\n")
	fmt.Fprintf(&responseString, "AI model   : %v%s\n", resp.ModelVersion, paramsStr)

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
	if progConfig.GeminiGroundingWithGoogleMaps {
		activeTools = append(activeTools, "Google Maps")
	}
	if len(includeStores) > 0 {
		activeTools = append(activeTools, "FileSearchStores")
	}

	if len(activeTools) > 0 {
		fmt.Fprintf(&responseString, "Tools      : %s\n", strings.Join(activeTools, ", "))
	}

	// slug extraction
	slug := "unknown-content"
	if len(resp.Candidates) > 0 {
		thoughts, content := getCandidateText(resp.Candidates[0])
		_, extractedSlug := extractAndCleanSlug(thoughts + "\n" + content)
		if extractedSlug != "" {
			slug = extractedSlug
		}
	}
	fmt.Fprintf(&responseString, "Slug       : %v\n", slug)

	fmt.Fprintf(&responseString, "Generated  : %v\n", finishProcessing.Format(time.RFC850))

	duration := finishProcessing.Sub(startProcessing)
	fmt.Fprintf(&responseString, "Processing : %.1f secs for %d %s\n", duration.Seconds(),
		len(resp.Candidates), pluralize(len(resp.Candidates), "candidate"))

	if resp.UsageMetadata != nil {
		u := resp.UsageMetadata
		fmt.Fprintf(&responseString, "Tokens     : %d (Total)\n", u.TotalTokenCount)
		totalInputCount := u.PromptTokenCount + u.ToolUsePromptTokenCount
		netPromptCount := u.PromptTokenCount - u.CachedContentTokenCount
		if netPromptCount < 0 {
			netPromptCount = 0
		}

		inputDetails := []string{fmt.Sprintf("Prompt: %d", netPromptCount)}
		if u.ToolUsePromptTokenCount > 0 {
			inputDetails = append(inputDetails, fmt.Sprintf("Tools: %d", u.ToolUsePromptTokenCount))
		}
		if u.CachedContentTokenCount > 0 {
			inputDetails = append(inputDetails, fmt.Sprintf("Cached: %d", u.CachedContentTokenCount))
		}
		fmt.Fprintf(&responseString, "  Input    : %d (%s)\n",
			totalInputCount, strings.Join(inputDetails, ", "))

		totalOutputCount := u.CandidatesTokenCount + u.ThoughtsTokenCount
		outputDetails := []string{fmt.Sprintf("Candidates: %d", u.CandidatesTokenCount)}
		if u.ThoughtsTokenCount > 0 {
			outputDetails = append(outputDetails, fmt.Sprintf("Thoughts: %d", u.ThoughtsTokenCount))
		}
		fmt.Fprintf(&responseString, "  Output   : %d (%s)\n",
			totalOutputCount, strings.Join(outputDetails, ", "))
	}

	if resp.PromptFeedback != nil {
		fmt.Fprintf(&responseString, "Blocked    : %v\n", resp.PromptFeedback.BlockReasonMessage)
	}

	responseString.WriteString("```\n")
	responseString.WriteString("<!-- STATS_END -->\n")

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

	// print response metadata
	responseString.WriteString("<!-- STATS_START -->\n")
	responseString.WriteString("```plaintext\n")
	if err == nil {
		fmt.Fprintf(&responseString, "AI model   : %v\n", progConfig.GeminiAiModel)
	}

	fmt.Fprintf(&responseString, "Slug       : error-response\n")

	fmt.Fprintf(&responseString, "Generated  : %v\n", finishProcessing.Format(time.RFC850))

	duration := finishProcessing.Sub(startProcessing)
	fmt.Fprintf(&responseString, "Processing : %.1f secs resulting in error\n", duration.Seconds())

	responseString.WriteString("```\n")
	responseString.WriteString("<!-- STATS_END -->\n")

	// append response string to request/response files
	appendResponseString(responseString)
}

/*
appendResponseString appends a given response string (which can be a successful response or an error message)
to the current request / response files in Markdown, ANSI, and HTML formats.
*/
func appendResponseString(responseString strings.Builder) {
	rawMarkdown := responseString.String()

	// extraxt Metadata Slug
	cleanedContent, _ := extractAndCleanSlug(rawMarkdown)

	// cleanup Markdown
	cleanedMarkdown := cleanMarkdown(cleanedContent)

	// 1. prepare Markdown for direct file saving (and ANSI rendering)
	// replace HTML comment tags with pure Markdown equivalents
	markdownForFileAndAnsi := cleanedMarkdown
	markdownForFileAndAnsi = strings.ReplaceAll(markdownForFileAndAnsi, "<!-- THOUGHTS_START -->", "**Thoughts:**\n")
	markdownForFileAndAnsi = strings.ReplaceAll(markdownForFileAndAnsi, "<!-- THOUGHTS_END -->", "")
	markdownForFileAndAnsi = strings.ReplaceAll(markdownForFileAndAnsi, "<!-- STATS_START -->", "**Statistics:**\n")
	markdownForFileAndAnsi = strings.ReplaceAll(markdownForFileAndAnsi, "<!-- STATS_END -->", "")
	markdownForFileAndAnsi = strings.ReplaceAll(markdownForFileAndAnsi, "<!-- RESPONSE_START -->", "**Response:**\n")
	markdownForFileAndAnsi = strings.ReplaceAll(markdownForFileAndAnsi, "<!-- RESPONSE_END -->", "")

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

	// 3. render markdown response as html (using cleaned string with comments)
	htmlData := cleanedMarkdown
	if progConfig.HTMLRendering {
		htmlData = renderMarkdown2HTML(cleanedMarkdown)
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
