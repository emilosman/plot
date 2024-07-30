package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	jsoniter "github.com/json-iterator/go"
)

type MarkdownFile struct {
	Date                     string  `json:"date"`
	Content                  string  `json:"content"`
	Diary                    string  `json:"diary"`
	TaskCompletionPercentage float64 `json:"taskCompletionPercentage"`
	WorkTime                 float64 `json:"workTime"`
	WeightVolume             int     `json:"weightVolume"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	directory := filepath.Join(os.Getenv("HOME"), "ruby", "exo", "daily")
	filesMap, err := readMarkdownFiles(directory)
	if err != nil {
		fmt.Printf("Error reading markdown files: %v\n", err)
		return
	}

	dates := make([]string, 0, len(filesMap))
	for date := range filesMap {
		dates = append(dates, date)
	}

	sort.Strings(dates)

	sortedFiles := make([]MarkdownFile, 0, len(filesMap))
	for _, date := range dates {
		content := filesMap[date]
		sortedFiles = append(sortedFiles, MarkdownFile{
			Date:                     date,
			Content:                  content,
			Diary:                    extractMarkdownSegment(content, "Diary"),
			TaskCompletionPercentage: taskCompletionPercentage(content),
			WorkTime:                 parseWorkTime(content),
			WeightVolume:             calculateTotalWorkload(extractMarkdownSegment(content, "Exercise")),
		})
	}

	w.Header().Set("Content-Type", "application/json")

	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	if err := json.NewEncoder(w).Encode(sortedFiles); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding JSON: %v", err), http.StatusInternalServerError)
		return
	}
}

func main() {
	http.HandleFunc("/", handler)
	fmt.Println("Server is listening on port 8080")
	http.ListenAndServe(":8080", nil)
}

func readMarkdownFiles(directory string) (map[string]string, error) {
	filesMap := make(map[string]string)

	files, err := ioutil.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("Failed to read directory: %w", err)
	}

	for _, file := range files {
		if file.Mode().IsRegular() && strings.HasSuffix(file.Name(), ".md") {
			fileNameParts := strings.Split(file.Name(), "-")
			if fileNameParts[1] == "template" || len(fileNameParts) < 2 {
				continue
			}
			datePart := fileNameParts[0]

			content, err := ioutil.ReadFile(filepath.Join(directory, file.Name()))
			if err != nil {
				return nil, fmt.Errorf("Failed to read file %s: %w", file.Name(), err)
			}

			filesMap[datePart] = string(content)
		}
	}

	return filesMap, nil
}

func taskCompletionPercentage(markdownContent string) float64 {
	uncheckedBox := regexp.MustCompile(`-\s*\[\s*\]`)
	checkedBox := regexp.MustCompile(`-\s*\[x\]`)

	uncheckedMatches := uncheckedBox.FindAllStringIndex(markdownContent, -1)
	checkedMatches := checkedBox.FindAllStringIndex(markdownContent, -1)

	totalCheckboxes := len(uncheckedMatches) + len(checkedMatches)
	checkedCheckboxes := len(checkedMatches)

	if totalCheckboxes == 0 {
		return 0.0
	}
	return (float64(checkedCheckboxes) / float64(totalCheckboxes)) * 100
}

func extractMarkdownSegment(markdownContent string, heading string) string {
	headingPattern := regexp.MustCompile(`(?m)^##\s+`)
	targetHeadingPattern := regexp.MustCompile(fmt.Sprintf(`(?m)^##\s+%s\s*$`, regexp.QuoteMeta(heading)))

	loc := targetHeadingPattern.FindStringIndex(markdownContent)
	if loc == nil {
		return ""
	}

	startIndex := loc[1]

	subsequentHeadings := headingPattern.FindAllStringIndex(markdownContent[startIndex:], 1)

	endIndex := len(markdownContent)
	if len(subsequentHeadings) > 0 {
		endIndex = startIndex + subsequentHeadings[0][0]
	}

	segment := markdownContent[startIndex:endIndex]

	return strings.TrimSpace(segment)
}

func calculateTotalWorkload(markdownContent string) int {
	lines := strings.Split(markdownContent, "\n")
	total := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		line = strings.TrimPrefix(line, "- ")

		parts := strings.Split(line, ",")
		if len(parts) < 3 {
			return 0
		}

		repsStr := strings.TrimSpace(parts[1])
		weightStr := strings.TrimSpace(parts[2])

		reps, err := strconv.Atoi(strings.TrimSuffix(repsStr, " reps"))
		if err != nil {
			return 0
		}

		weight := 0
		if weightStr == "bodyweight" {
			weight = 10
		} else {
			weight, err = strconv.Atoi(strings.TrimSuffix(weightStr, "kg"))
			if err != nil {
				return 0
			}
		}

		total += reps * weight
	}

	return total
}

func parseWorkTime(markdownContent string) float64 {
	workTimePattern := regexp.MustCompile(`work:: (\d{2}):(\d{2})`)

	matches := workTimePattern.FindAllStringSubmatch(markdownContent, -1)
	if matches == nil {
		return 0
	}

	totalHours := 0.0

	for _, match := range matches {
		if len(match) != 3 {
			continue
		}

		hours, err := strconv.Atoi(match[1])
		if err != nil {
			return 0
		}

		minutes, err := strconv.Atoi(match[2])
		if err != nil {
			return 0
		}

		timeInHours := float64(hours) + float64(minutes)/60.0
		totalHours += timeInHours
	}

	return totalHours
}
