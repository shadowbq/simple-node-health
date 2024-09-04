package commonutils

import (
	"encoding/json"
	"strings"
)

// removeDuplicates removes duplicate values from a slice
func removeDuplicatesFromSlice(slice []string) []string {
	seen := make(map[string]struct{})
	uniqueSlice := []string{}

	for _, item := range slice {
		if _, found := seen[item]; !found {
			seen[item] = struct{}{}
			uniqueSlice = append(uniqueSlice, item)
		}
	}

	return uniqueSlice
}

// Function to convert multi-line string to JSON object
func multiLineStringToJSON(input string) (string, error) {
	// Split the input string by newlines to create a slice of strings
	lines := strings.Split(strings.TrimSpace(input), "\n")

	// Create a map to hold the JSON object
	jsonObject := map[string][]string{
		"response": lines,
	}

	// Convert the map to a JSON string
	jsonBytes, err := json.MarshalIndent(jsonObject, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// Function to convert a []string to a JSON object
func stringArrayToJSON(input []string) (string, error) {
	// Create a map with the key "response" and value as the input slice
	jsonObject := map[string][]string{
		"response": input,
	}

	// Convert the map to a JSON string with indentation
	jsonBytes, err := json.MarshalIndent(jsonObject, "", "  ")
	if err != nil {
		return "", err
	}

	// Return the JSON string
	return string(jsonBytes), nil
}
