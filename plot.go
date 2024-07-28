package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func handler(w http.ResponseWriter, r *http.Request) {
	directory := filepath.Join(os.Getenv("HOME"), "ruby", "exo", "daily")
	filesMap, err := readMarkdownFiles(directory)
	if err != nil {
		fmt.Printf("Error reading markdown files: %v\n", err)
		return
	}

	// Print the map
	for date, content := range filesMap {
		fmt.Fprintf(w, "Date: %s\nContent:\n%s\n\n", date, content)
	}
}

func main() {
	http.HandleFunc("/", handler)
	fmt.Println("Server is listening on port 8080")
	http.ListenAndServe(":8080", nil)
}

func readMarkdownFiles(directory string) (map[string]string, error) {
	filesMap := make(map[string]string)

	// Read the directory
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("Failed to read directory: %w", err)
	}

	// Loop through the files
	for _, file := range files {
		// Check if the file is a regular file and has the .md extension
		if file.Mode().IsRegular() && strings.HasSuffix(file.Name(), ".md") {
			// Extraxt date part of filename
			fileNameParts := strings.Split(file.Name(), "-")
			if len(fileNameParts) < 2 {
				continue // skip files that do not match the expected pattern
			}
			datePart := fileNameParts[0]

			// Read the file content
			content, err := ioutil.ReadFile(filepath.Join(directory, file.Name()))
			if err != nil {
				return nil, fmt.Errorf("Failed to read file %s: %w", file.Name(), err)
			}

			// Store the contents in the map
			filesMap[datePart] = string(content)
		}
	}

	return filesMap, nil
}
