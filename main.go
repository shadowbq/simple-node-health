package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	Version string // set by the build process
)

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

func getStatus() map[string]string {
	return map[string]string{"status": "ok"}
}

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

	jsonOutput, err := stringArrayToJSON(results)
	if err != nil {
		return "", fmt.Errorf("Error: %v", err)
	}

	return jsonOutput, nil
}

func getDNS(domain string) (string, error) {
	cmd := exec.Command("dig", "+short", domain)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Error executing dig command: %v", err)
	}

	jsonOutput, err := multiLineStringToJSON(string(output))

	if err != nil {
		return "", fmt.Errorf("Error: %v", err)
	}

	return string(jsonOutput), nil
}

// Function to return JSON status
func checkStatus(w http.ResponseWriter, r *http.Request) {
	response := getStatus()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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

// Start the web server with configurable port
func startServer(port int) {
	http.HandleFunc("/check", checkStatus)
	http.HandleFunc("/check/disks", checkDisks)
	http.HandleFunc("/check/dns", checkDNS)

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("Starting server on %s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
		os.Exit(1)
	}
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

// Function to print check disks to console
func runCheckDisks(cmd *cobra.Command, args []string) {
	response, err := getDisks()
	if err != nil {
		fmt.Println("Error checking Disks:", err)
		return
	}
	fmt.Print(response)
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

func main() {
	var port int
	var domain string
	var verbose bool

	// Root command
	var rootCmd = &cobra.Command{
		Use:   "simple-node-health",
		Short: "A simple tool to check hardware EXT4 devices and run DNS queries",
		Run: func(cmd *cobra.Command, args []string) {
			startServer(port)
		},
	}

	// settingsCmd represents the settings command
	var settingsCmd = &cobra.Command{
		Use:   "settings",
		Short: "Print the current configuration settings",
		Long: `Print the current configuration settings. This command is useful to see the final configuration once all the settings have been applied. 
		It also shows how to access the global flags and command flags.`,
		Run: func(cmd *cobra.Command, args []string) {
			//fmt.Println("Settings called")

			if verbose {
				fmt.Println("--- Final configuration  ---")
			}
			//fmt.Printf("\tVersion: %s\n", Version)
			//for s, i := range viper.AllSettings() {
			//	fmt.Printf("\t%s: %v\n", s, i)
			//}

			keys := viper.AllSettings()
			//sort keys
			var keysSorted []string
			for key := range keys {
				keysSorted = append(keysSorted, key)
			}

			// get the keys and print them in sorted order
			for _, key := range keysSorted {
				if key == "clientsecret" {
					// Print the clientSecret first 4 characters then the rest as *
					fmt.Printf("\t%s: %v\n", key, viper.Get(key).(string)[:4]+"********")
				} else {
					fmt.Printf("\t%s: %v\n", key, viper.Get(key))
				}
			}

			if verbose {
				fmt.Println("----------------------------")
			}

		},
	}

	rootCmd.AddCommand(settingsCmd)

	// Check command
	var checkCmd = &cobra.Command{
		Use:   "check",
		Short: "Run various checks",
	}

	// Subcommand: checkstatus
	var checkStatusCmd = &cobra.Command{
		Use:   "status",
		Short: "Check the service status",
		Run:   runCheckStatus,
	}

	// Subcommand: checkdisks
	var checkDisksCmd = &cobra.Command{
		Use:   "disks",
		Short: "Check EXT4 devices for read-only mode",
		Run:   runCheckDisks,
	}

	// Subcommand: checkdns
	var checkDNSCmd = &cobra.Command{
		Use:   "dns",
		Short: "Run a DNS query for the specified domain",
		Run:   runCheckDNS,
	}

	// Add subcommands to the check command
	checkCmd.AddCommand(checkStatusCmd, checkDisksCmd, checkDNSCmd)

	// Add the check command to the root command
	rootCmd.AddCommand(checkCmd)

	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "verbose output")
	// bind the configuration to file/environment variables
	cobra.CheckErr(viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose")))
	viper.SetDefault("verbose", false)

	// Domain flag
	rootCmd.PersistentFlags().StringVarP(&domain, "domain", "d", "cloudflare.com", "Domain to query with dig")
	viper.BindPFlag("domain", rootCmd.PersistentFlags().Lookup("domain"))

	// Port flag
	rootCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port for the web server")
	viper.BindPFlag("port", rootCmd.Flags().Lookup("port"))

	// Bind environment variables
	viper.AutomaticEnv()

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
