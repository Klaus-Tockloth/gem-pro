package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

/*
readPromptFromKeyboard reads user prompts from the standard input (keyboard/stdin). It continuously reads
lines from stdin, treating each line as a prompt, and sends them to the prompt channel for processing.
*/
func readPromptFromKeyboard(promptChannel chan string) {
	reader := bufio.NewReader(os.Stdin)
	for {
		promptData, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("error [%v] at reader.ReadString()", err)
			return
		}
		if promptData == "\n" || promptData == "\r\n" {
			continue
		}

		// read prompt from given text file (e.g. "<<<MyQuery.txt" or "<<< MyQuery.txt")
		var fileData []byte
		if strings.HasPrefix(promptData, "<<<") {
			filename := strings.TrimSpace(strings.TrimPrefix(promptData, "<<<"))
			fileData, err = os.ReadFile(filename)
			if err != nil {
				fmt.Printf("error [%v] at os.ReadFile()\n", err)
				continue
			}
			if len(fileData) > 0 {
				promptChannel <- string(fileData)
			}
		} else {
			promptChannel <- promptData
		}
	}
}

/*
readPromptFromFile monitors a specified file for changes and reads its content as a prompt. It watches a
given file for modifications in size or modification time, and upon change, reads the file content and
sends it to the prompt channel.
*/
func readPromptFromFile(filePath string, promptChannel chan string) {
	currentStat, err := os.Stat(filePath)
	if err != nil {
		fmt.Printf("error [%v] at os.Stat()", err)
	}
	for {
		stat, err := os.Stat(filePath)
		if err != nil {
			fmt.Printf("error [%v] at os.Stat()", err)
		}
		if stat.Size() != currentStat.Size() || stat.ModTime() != currentStat.ModTime() {
			promptData, err := os.ReadFile(filePath)
			if err != nil {
				fmt.Printf("error [%v] at os.ReadFile()", err)
			}
			if len(promptData) > 0 {
				promptChannel <- string(promptData)
			}
			currentStat = stat
		}
		time.Sleep(500 * time.Millisecond)
	}
}

/*
readPromptFromLocalhost creates an HTTP handler function to receive prompts from localhost. It sets up an
HTTP handler that listens for POST requests on localhost, reads the request body as a prompt, and sends it
through the prompt channel.
*/
func readPromptFromLocalhost(promptChannel chan string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "error reading request body", http.StatusBadRequest)
			fmt.Printf("error [%v] reading request body\n", err)
			return
		}

		if len(body) == 0 {
			http.Error(w, "prompt empty", http.StatusBadRequest)
			return
		}
		promptChannel <- string(body)
		defer r.Body.Close()

		fmt.Fprintln(w, "prompt received")
	}
}
