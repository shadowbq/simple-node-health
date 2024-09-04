package parsers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"

	_ "github.com/shadowbq/simple-node-health/commonutils"
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

	jsonOutput, err := commonutils.stringArrayToJSON(results)
	if err != nil {
		return "", fmt.Errorf("Error: %v", err)
	}

	return jsonOutput, nil
}

// Function to check for EXT4 devices in read-only mode
func checkDisks(w http.ResponseWriter, r *http.Request) {
	results, err := getDisks()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error checking disks: %v\n", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// Function to print check disks to console
func runCheckDisks(cmd *cobra.Command, args []string) {
	response, err := getDisks()
	if err != nil {
		fmt.Println("Error checking Disks:", err)
		return
	}
	fmt.Print(response)
}
