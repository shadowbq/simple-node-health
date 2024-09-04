package parsers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

func getStatus() map[string]string {
	return map[string]string{"status": "ok"}
}

// Function to return JSON status
func checkStatus(w http.ResponseWriter, r *http.Request) {
	response := getStatus()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Function to print check status to console
func runCheckStatus(cmd *cobra.Command, args []string) {
	response := getStatus()
	output, err := json.Marshal(response)
	if err != nil {
		fmt.Println("Error encoding response:", err)
		return
	}
	fmt.Println(string(output))
}
