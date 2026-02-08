package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/davecgh/go-spew/spew"
	"github.com/gabriel-vasile/mimetype"
	"github.com/gofrs/uuid"
	"github.com/mitchellh/go-wordwrap"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var (
	metadataSlugRegex = regexp.MustCompile(`(?m)^METADATA_SLUG:\s*(.+)\s*$`)
	slugAllowedChars  = regexp.MustCompile(`[^a-z0-9-]+`)
	slugMultiDash     = regexp.MustCompile(`-+`)
)

/*
extractAndCleanSlug extracts the slug from the content and returns the cleaned content
as well as the found slug.
If no slug is found, the content is returned unchanged and an empty string is returned.
*/
func extractAndCleanSlug(content string) (string, string) {
	// Search for the slug
	matches := metadataSlugRegex.FindStringSubmatch(content)
	slug := ""

	if len(matches) > 1 {
		rawSlug := matches[1]
		slug = sanitizeSlug(rawSlug)

		// Remove the metadata line from the content so it does not appear in the output
		content = metadataSlugRegex.ReplaceAllString(content, "")
		content = strings.TrimSpace(content)
	}

	return content, slug
}

/*
sanitizeSlug ensures that the slug conforms to the "lowercase, ascii only, kebab-case" format.
It uses generic unicode normalization to handle chars like ñ, ç, é, etc.
*/
func sanitizeSlug(input string) string {
	// Basic Cleaning and Lowercasing
	input = strings.TrimSpace(input)
	input = strings.ToLower(input)

	// Specific Cultural Overrides (Business Logic)
	// NFD normalization would turn "ä" -> "a", but we explicitly want "ae".
	// Therefore, we handle these German specifics BEFORE generic normalization.
	input = strings.ReplaceAll(input, "ä", "ae")
	input = strings.ReplaceAll(input, "ö", "oe")
	input = strings.ReplaceAll(input, "ü", "ue")
	input = strings.ReplaceAll(input, "ß", "ss")

	// Generic Unicode Normalization (NFD + Remove Non-Spacing Marks)
	// This handles cases like:
	// "ñ" -> "n" + "~" -> "n"
	// "é" -> "e" + "´" -> "e"
	// "ç" -> "c" + "¸" -> "c"
	t := transform.Chain(
		norm.NFD,                           // Decompose characters
		runes.Remove(runes.In(unicode.Mn)), // Remove non-spacing marks (accents, tildes, etc.)
		norm.NFC,                           // Recompose (optional, ensures standard form for remaining chars)
	)

	result, _, err := transform.String(t, input)
	if err != nil {
		// Fallback to original input if transformation fails (unlikely)
		fmt.Printf("warning: slug normalization failed for '%s': %v\n", input, err)
		result = input
	}
	input = result

	// Sanitize: Keep only alphanumeric ASCII and hyphens
	// This removes any characters that survived normalization but aren't letters/numbers
	// (e.g. emojis, punctuation like '?', '!', brackets).
	input = slugAllowedChars.ReplaceAllString(input, "-")

	// Cleanup Dashes
	// No double hyphens
	input = slugMultiDash.ReplaceAllString(input, "-")
	// Trim dashes from start/end
	input = strings.Trim(input, "-")

	// Truncate
	if len(input) > 100 {
		input = input[:100]
		input = strings.TrimRight(input, "-")
	}

	return input
}

/*
buildDestinationFilename generates the filename according to the schema:
yyyymmdd-hhmmss-<slug>.<extension>
*/
func buildDestinationFilename(now time.Time, slug string, extension string) string {
	const fallbackSlug = "unknown-content"

	if slug == "" {
		slug = fallbackSlug
	}
	timestamp := now.Format("20060102-150405")
	filename := fmt.Sprintf("%s-%s.%s", timestamp, slug, extension)
	return filename
}

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
dirExists checks if a directory exists at the given path. It returns true if the path exists and
is a directory, otherwise false.
*/
func dirExists(dir string) bool {
	info, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return false // path does not exist
	}
	if err != nil {
		// another error occurred (e.g., permissions): for simplicity, we treat any other error as "does not exist or cannot check"
		return false
	}
	return info.IsDir() // path exists, check if it's a directory
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

	err = os.WriteFile(destinationFile, input, 0600)
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
runCommand executes an external command line command or program. It takes a command line string,
parses it into a command and its arguments using `splitCommandLine` (handling quoted arguments),
and then executes the command using `os/exec.Command().Run()`. If the command execution fails,
it prints an error message to standard output and returns the error.
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
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// ignore empty lines and lines starting with # or //
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}
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
dumpDataToFile writes an arbitrary data object to a file in a human-readable format using `spew.Sdump`. It
serializes any given Go data object into a human-readable string format using `spew.Sdump` and writes this
string to a file, useful for debugging and logging purposes.
*/
func dumpDataToFile(flag int, objectname string, object interface{}) {
	data := fmt.Sprintf("---------- %s ----------\n%s\n", objectname, spew.Sdump(object))
	file, err := os.OpenFile("gemini.raw", flag, 0600)
	if err != nil {
		fmt.Printf("error [%v] at os.OpenFile()\n", err)
		return
	}
	defer func() { _ = file.Close() }()
	_, _ = fmt.Fprint(file, data)
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
	err = os.WriteFile(pathname, data, 0600)
	if err != nil {
		return "", "", fmt.Errorf("error [%v] at os.WriteFile()", err)
	}

	return pathname, filename, nil
}

/*
parseMIMETypeReplacements parses a list of strings containing MIME type pairs
(e.g., "key1 = value1", "key2 = value2") and returns them as a map[string]string.
It returns an error if a pair is malformed.
*/
func parseMIMETypeReplacements(replacements []string) (map[string]string, error) {
	mimeMap := make(map[string]string)

	for _, pair := range replacements {
		trimmedPair := strings.TrimSpace(pair)

		// Ignore empty entries that might occur due to additional newlines.
		if trimmedPair == "" {
			continue
		}

		// Split the pair into key and value at the first '='.
		parts := strings.SplitN(trimmedPair, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid MIME type pair format: %q. Expected 'key = value'", trimmedPair)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if key == "" {
			return nil, fmt.Errorf("MIME type key cannot be empty in pair: %q", trimmedPair)
		}
		mimeMap[key] = value
	}

	return mimeMap, nil
}
