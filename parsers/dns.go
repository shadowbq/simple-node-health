package parsers

import (
	"fmt"
	"net/http"
	"os/exec"

	"github.com/shadowbq/simple-node-health/helpers"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Function to run `dig <domain>` and return the result
func getDNS(domain string) (string, error) {
	cmd := exec.Command("dig", "+short", domain)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Error executing dig command: %v", err)
	}

	jsonOutput, err := helpers.MultiLineStringToJSON(string(output))

	if err != nil {
		return "", fmt.Errorf("Error: %v", err)
	}

	return string(jsonOutput), nil
}

// Function to call getDNS and return the result
func HTTPCheckDNS(w http.ResponseWriter, r *http.Request) {
	domain := viper.GetString("domain")
	result, err := getDNS(domain)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error checking DNS: %v\n", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintln(w, result)
}

// Function to print check DNS to console
func CmdCheckDNS(cmd *cobra.Command, args []string) {
	domain := viper.GetString("domain")
	response, err := getDNS(domain)
	if err != nil {
		fmt.Println("Error checking DNS:", err)
		return
	}
	fmt.Print(response)
}
