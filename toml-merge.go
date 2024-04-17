package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
)

var (
	flagOutputJSON  = false
	flagShowVersion = false
	version         = "99.99-local"
)

func init() {
	log.SetFlags(0)

	flag.BoolVar(&flagOutputJSON, "output-json", flagOutputJSON, "Output JSON instead of TOML")
	flag.BoolVar(&flagShowVersion, "version", flagShowVersion, "Show version")

	flag.Usage = usage
	flag.Parse()
}

func usage() {
	log.Printf("Usage: %s toml-file [ toml-file ... ]\n", path.Base(os.Args[0]))
	flag.PrintDefaults()

	os.Exit(1)
}

func readLinesFromFile(filePath string) ([]string, error) {
	// Read the contents of the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Split the content into lines
	lines := strings.Split(string(content), "\n")

	// Remove empty lines
	var nonEmptyLines []string
	for _, line := range lines {
		if line != "" {
			nonEmptyLines = append(nonEmptyLines, line)
		}
	}

	return nonEmptyLines, nil
}

// removeDuplicates removes duplicate strings from the given slice.
func removeDuplicates(strs []string) []string {
	encountered := map[string]bool{} // Map to store encountered strings
	result := []string{}             // Slice to store unique strings

	// Iterate over the slice
	for _, str := range strs {
		// If the string has not been encountered before, add it to the result slice
		if !encountered[str] {
			result = append(result, str)
			encountered[str] = true // Mark the string as encountered
		}
	}

	return result
}

func main() {
	if flagShowVersion {
		log.Printf("toml-merge version %s\n", version)
		os.Exit(0)
	}

	if flag.NArg() < 1 {
		flag.Usage()
	}

	var files = flag.Args()

	// Define the regular expression pattern
	pattern := `^\[.*\]`
	// Compile the regular expression pattern
	regex, err := regexp.Compile(pattern)
	if err != nil {
		fmt.Println("Error compiling regex:", err)
		return
	}

	var ouputMap = make(map[string][]string)
	outputCommentsAndPrimary := make([]string, 0)

	for _, filePath := range files {
		lines, err := readLinesFromFile(filePath)
		if err != nil {
			fmt.Printf("Error reading file '%s': %v\n", filePath, err)
			continue
		}

		// keep track of current heading
		var currentSection string

		for _, line := range lines {
			if regex.MatchString(line) {
				currentSection = line
				if _, ok := ouputMap[currentSection]; !ok {
					ouputMap[currentSection] = []string{}
				}
				continue
			}
			if currentSection != "" {
				ouputMap[currentSection] = append(ouputMap[currentSection], line)
			}
			if currentSection == "" {
				outputCommentsAndPrimary = append(outputCommentsAndPrimary, line)
			}
		}
	}

	// Remove duplicates from the output
	for key, value := range ouputMap {
		ouputMap[key] = removeDuplicates(value)
	}

	// Flatten the map into one string with newlines
	var flattened strings.Builder
	for _, value := range outputCommentsAndPrimary {
		// flattened.WriteString(key)
		// flattened.WriteString(": ")
		flattened.WriteString(value)
		flattened.WriteString("\n")
	}
	flattened.WriteString("\n")

	for key, value := range ouputMap {
		flattened.WriteString(key)
		flattened.WriteString("\n")
		for _, value := range value {
			flattened.WriteString(value)
			flattened.WriteString("\n")
		}
		flattened.WriteString("\n")
	}

	// Convert the flattened string to a string
	result := flattened.String()
	fmt.Println(result)

	// if flagOutputJSON {
	// 	jsonwrt := json.NewEncoder(os.Stdout)
	// 	jsonwrt.SetIndent("", "  ")
	// 	jsonwrt.Encode(output)
	// } else {
	// 	tomlrt := toml.NewEncoder(os.Stdout)
	// 	tomlrt.Encode(output)
	// }
}
