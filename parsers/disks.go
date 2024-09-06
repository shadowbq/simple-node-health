package parsers

import (
	"fmt"
	"net/http"
	"os/exec"
	"strings"

	"github.com/shadowbq/simple-node-health/helpers"
	"github.com/spf13/cobra"
)

func getDisks() (string, error) {
	cmd := exec.Command("mount")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Error executing mount command: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	var results []string

	for _, line := range lines {
		if strings.Contains(line, "type ext4") && strings.Contains(line, "ro,") {
			results = append(results, line)
		}
	}

	// Check if the slice is nil and assign [""] if it is
	if results == nil {
		results = []string{""}
	}

	//return results, nil

	jsonOutput, err := helpers.StringArrayToJSON(results)
	if err != nil {
		return "", fmt.Errorf("Error: %v", err)
	}

	return string(jsonOutput), nil
}

// Function to check for EXT4 devices in read-only mode
func HTTPCheckDisks(w http.ResponseWriter, r *http.Request) {
	result, err := getDisks()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error checking disks: %v\n", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	//json.NewEncoder(w).Encode(results)
	fmt.Fprintln(w, result)
}

// Function to print check disks to console
func CmdCheckDisks(cmd *cobra.Command, args []string) {
	response, err := getDisks()
	if err != nil {
		fmt.Println("Error checking Disks:", err)
		return
	}
	fmt.Print(response)
}
