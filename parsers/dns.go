package parsers

import (
	"fmt"
	"net/http"
	"os/exec"

	_ "github.com/shadowbq/simple-node-health/commonutils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func getDNS(domain string) (string, error) {
	cmd := exec.Command("dig", "+short", domain)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Error executing dig command: %v", err)
	}

	jsonOutput, err := commonutils.multiLineStringToJSON(string(output))

	if err != nil {
		return "", fmt.Errorf("Error: %v", err)
	}

	return string(jsonOutput), nil
}

// Function to run `dig <domain>` and return the result
func checkDNS(w http.ResponseWriter, r *http.Request) {
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
func runCheckDNS(cmd *cobra.Command, args []string) {
	domain := viper.GetString("domain")
	response, err := getDNS(domain)
	if err != nil {
		fmt.Println("Error checking DNS:", err)
		return
	}
	fmt.Print(response)
}
