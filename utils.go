package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/flytam/filenamify"
	"github.com/gabriel-vasile/mimetype"
	"github.com/gofrs/uuid"
	"github.com/mitchellh/go-wordwrap"
)

/*
fileExists checks if a file exists at the given filename. It verifies whether a file exists at
the provided path and ensures it is not a directory.
*/
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

/*
wrapString wraps a long string to a specified width for better readability in terminal output. It takes a
long string and formats it by wrapping it to a specified width, inserting line breaks and indentation for
improved terminal display.
*/
func wrapString(message string, width int, ident int) string {
	wrapped := wordwrap.WrapString(message, uint(width-ident))
	wrapped = strings.ReplaceAll(wrapped, "\n", "\n"+strings.Repeat(" ", ident))
	return wrapped
}

/*
copyFile copies the content of a source file to a destination file. It reads all content from the source
file and writes it to the destination file, effectively duplicating the file content.
*/
func copyFile(sourceFile, destinationFile string) {
	input, err := os.ReadFile(sourceFile)
	if err != nil {
		fmt.Printf("error [%v] at os.ReadFile()\n", err)
		return
	}

	err = os.WriteFile(destinationFile, input, 0644)
	if err != nil {
		fmt.Printf("error [%v] at os.WriteFile()\n", err)
		return
	}
}

/*
pluralize adds a simple plural suffix "s" to a singular word if the count is not equal to 1. It conditionally
adds an "s" to a given singular word, creating a plural form based on whether a count is one or more than one.
*/
func pluralize(count int, singular string) string {
	if count == 1 {
		return singular
	}
	return singular + "s"
}

/*
runCommand executes a command line command or program.It takes a command line string, parses it into command
and arguments, and then executes it using 'os/exec'.
*/
func runCommand(commandLine string) error {
	parsedArgs := splitCommandLine(commandLine)
	cmd := exec.Command(parsedArgs[0], parsedArgs[1:]...)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("error [%v] executing command [%v]\n", err, commandLine)
	}
	return err
}

/*
splitCommandLine parses a command line string into a slice of strings, separating the command and its arguments.
It tokenizes a command line string, handling quoted arguments to correctly separate commands and their arguments
into a string slice.
*/
func splitCommandLine(commandLine string) []string {
	var args []string
	var inQuote bool
	var quoteType rune // ' or "
	var currentArg strings.Builder

	for _, r := range commandLine {
		switch {
		case r == '"' || r == '\'':
			if inQuote {
				if quoteType == r {
					inQuote = false
					args = append(args, currentArg.String())
					currentArg.Reset()
				} else {
					// inside a quotation mark a different type is found, so treat it as part of the argument
					currentArg.WriteRune(r)
				}
			} else {
				inQuote = true
				quoteType = r
			}
		case r == ' ' && !inQuote:
			if currentArg.Len() > 0 {
				args = append(args, currentArg.String())
				currentArg.Reset()
			}
		default:
			currentArg.WriteRune(r)
		}
	}

	// add remaining argument, if any
	if currentArg.Len() > 0 {
		args = append(args, currentArg.String())
	}

	return args
}

/*
getPassword retrieves a password from a passphrase string. It extracts a password from a passphrase
string, which can specify different sources for the password such as 'pass:', 'env:', or 'file:',
and handles retrieval based on the source.
*/
func getPassword(passPhrase string) (string, error) {
	items := strings.SplitN(passPhrase, ":", 2)
	source := strings.ToLower(items[0])
	password := ""

	if len(items) != 2 {
		return "", fmt.Errorf("unable to split pass phrase argument into 'source:password'")
	}

	switch source {
	case "pass":
		password = items[1]
	case "env":
		password = os.Getenv(items[1])
		if password == "" {
			return "", fmt.Errorf("password empty or env variable [%s] not found", items[1])
		}
	case "file":
		// read password file
		lines, err := slurpFile(items[1])
		if err != nil || len(lines) == 0 {
			return "", fmt.Errorf("unable to read password from file, error = [%w], file = [%v]", err, items[1])
		}
		password = lines[0]
	default:
		return "", fmt.Errorf("invalid password source (not 'pass:', 'env:' or 'file:')")
	}

	return password, nil
}

/*
slurpFile reads all lines from a text file and returns them as a slice of strings. It reads the content of a
text file line by line and returns each line as an element in a string slice.
*/
func slurpFile(filename string) ([]string, error) {
	var lines []string

	file, err := os.Open(filename)
	if err != nil {
		return lines, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
	}

	return lines, nil
}

/*
getFileMimeType detects the MIME type of a file based on its content. It analyzes a file's content to
determine its MIME type, providing a string representation of the detected type.
*/
func getFileMimeType(filepath string) (string, error) {
	mtp, err := mimetype.DetectFile(filepath)
	if err != nil {
		return "application/octet-stream", err
	}

	mimeType := mtp.String()

	// e.g. 'text/plain; charset=utf-8' -> 'text/plain'
	mimeTypeParts := strings.Split(mimeType, ";")
	mimeType = mimeTypeParts[0]

	return mimeType, nil
}

/*
promptToFilename generates a valid filename from a user prompt string. It creates a safe and valid
filename from a user prompt by replacing problematic characters, truncating it to a maximum length,
and adding prefixes, postfixes, and extensions as configured.
*/
func promptToFilename(prompt string, maxLength int, prefix, postfix, extension string) string {
	var filename string
	var err error

	// length correction for 'core' name
	maxLength -= 2 // "[" + "]"
	if prefix != "" {
		maxLength -= len(prefix) + 1
	}
	if postfix != "" {
		maxLength -= 1 + len(postfix)
	}
	if extension != "" {
		maxLength -= 1 + len(extension)
	}

	// replace problematic characters with visuell similar runes
	prompt = strings.ReplaceAll(prompt, "?", "ʔ")  // glottal stop
	prompt = strings.ReplaceAll(prompt, ":", "ː")  // triangular colon
	prompt = strings.ReplaceAll(prompt, "/", "∕")  // division slash
	prompt = strings.ReplaceAll(prompt, "\\", "＼") // fullwidth reverse solidus
	prompt = strings.ReplaceAll(prompt, "*", "⁎")  // low asterisk
	prompt = strings.ReplaceAll(prompt, "|", "¦")  // broken bar
	prompt = strings.ReplaceAll(prompt, "<", "‹")  // single left-pointing angle quotation mark
	prompt = strings.ReplaceAll(prompt, ">", "›")  // single right-pointing angle quotation mark
	prompt = strings.ReplaceAll(prompt, "\"", "”") // right double quotation mark
	prompt = strings.ReplaceAll(prompt, ".", "․")  // one dot leader

	filename, err = filenamify.Filenamify(prompt, filenamify.Options{Replacement: " ", MaxLength: maxLength})
	if err != nil {
		fmt.Printf("error [%v] at filenamify.Filenamify()\n", err)
		uuid4, _ := uuid.NewV4()
		filename = uuid4.String()
	}
	filename = "[" + filename + "]"

	if prefix != "" {
		filename = prefix + "." + filename
	}
	if postfix != "" {
		filename += "." + postfix
	}
	if extension != "" {
		filename += "." + extension
	}

	return filename
}

/*
buildDestinationFilename constructs a destination filename for history files based on the program configuration
and current context. It generates a filename for saving history files, based on the configured filename schema
(timestamp or prompt-based), adding prefixes, postfixes, and extensions as specified in the program settings.
*/
func buildDestinationFilename(now time.Time, prompt, extension string) string {
	formatLayout := "20060102-150405"
	timestamp := now.Format(formatLayout)

	destinationFilename := ""
	switch progConfig.HistoryFilenameSchema {
	case "prompt":
		prefix := ""
		if progConfig.HistoryFilenameAddPrefix {
			prefix = timestamp
		}
		postfix := ""
		if progConfig.HistoryFilenameAddPostfix {
			postfix = timestamp
		}
		destinationFilename = promptToFilename(prompt, progConfig.HistoryMaxFilenameLength, prefix, postfix, extension)
	case "timestamp":
		destinationFilename = timestamp
		if extension != "" {
			destinationFilename += "." + extension
		}
	}
	return destinationFilename
}

/*
dumpDataToFile writes an arbitrary data object to a file in a human-readable format using `spew.Sdump`. It
serializes any given Go data object into a human-readable string format using `spew.Sdump` and writes this
string to a file, useful for debugging and logging purposes.
*/
func dumpDataToFile(flag int, objectname string, object interface{}) {
	data := fmt.Sprintf("---------- %s ----------\n%s\n", objectname, spew.Sdump(object))
	file, err := os.OpenFile("gemini.raw", flag, 0644)
	if err != nil {
		fmt.Printf("error [%v] at os.OpenFile()\n", err)
		return
	}
	defer file.Close()
	fmt.Fprint(file, data)
}

/*
writeDataToFile writes the provided byte slice data to a file. It generates a unique filename based on the
current timestamp, a UUID, and the provided mimeType. It saves the file in the "files" directory. It returns
the full path and the filename to the written file or an error if any step fails.
*/
func writeDataToFile(data []byte, mimeType string, timestamp time.Time) (string, string, error) {
	// generate UUID4
	uuid4, err := uuid.NewV4()
	if err != nil {
		return "", "", fmt.Errorf("error [%v] at uuid.NewV4()", err)
	}

	// modify MIME type separator
	mimeType = strings.ReplaceAll(mimeType, "/", ".")

	// create directory
	directory := "files"
	err = os.MkdirAll(directory, 0750)
	if err != nil {
		return "", "", fmt.Errorf("error [%v] at os.MkdirAll()", err)
	}

	// get working directory
	wd, err := os.Getwd()
	if err != nil {
		return "", "", fmt.Errorf("error [%v] at os.Getwd()", err)
	}

	// build unique filename
	filename := timestamp.Format("20060102-150405") + "-" + uuid4.String() + "." + mimeType

	// build absolute path
	pathname := filepath.Join(wd, directory, filename)

	// write file
	err = os.WriteFile(pathname, data, 0644)
	if err != nil {
		return "", "", fmt.Errorf("error [%v] at os.WriteFile()", err)
	}

	return pathname, filename, nil
}
