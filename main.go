package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	Version string // set by the build process
)

// Create a logger
var auditLogger *log.Logger

var routes []string

type Client struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
}

type Claims struct {
	ClientID string `json:"client_id"`
	jwt.RegisteredClaims
}

var clients []Client
var authTokenSecret string

var mux *RouteTrackingMux

var mainMux *RouteTrackingMux

//var secureMux *http.ServeMux

func initAuditLogger() {
	file, err := os.OpenFile("audit.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open audit log file: %v", err)
	}

	auditLogger = log.New(file, "", log.LstdFlags)
}

// Log a message to the audit log
func auditLog(message string) {
	auditLogger.Println(message)
}

type RouteTrackingMux struct {
	*http.ServeMux
	routes []string
}

func NewRouteTrackingMux() *RouteTrackingMux {
	return &RouteTrackingMux{
		ServeMux: http.NewServeMux(),
		routes:   make([]string, 0),
	}
}

func (rtm *RouteTrackingMux) Handle(pattern string, handler http.Handler) {
	rtm.routes = append(rtm.routes, pattern)
	rtm.ServeMux.Handle(pattern, handler)
}

func (rtm *RouteTrackingMux) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	rtm.routes = append(rtm.routes, pattern)
	rtm.ServeMux.HandleFunc(pattern, handler)
}

func (rtm *RouteTrackingMux) Routes() []string {
	return rtm.routes
}

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

// Function GenerateJWT creates a new JWT token
func generateJWT(clientID string) (string, error) {
	// Define your claims
	claims := jwt.MapClaims{
		"client_id": clientID,
		"exp":       time.Now().Add(time.Hour).Unix(), // Token expires in 1 hour
	}

	// Create a new JWT token with the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret
	tokenString, err := token.SignedString([]byte(authTokenSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// Function to validate client credentials
func validateClientCredentials(clientID, clientSecret string) bool {
	for _, client := range clients {
		if client.ClientID == clientID && client.ClientSecret == clientSecret {
			return true
		}
	}
	return false
}

// Function to handle Token and call generator for JWT token
func tokenHandler(w http.ResponseWriter, r *http.Request) {
	clientID := r.FormValue("client_id")
	clientSecret := r.FormValue("client_secret")

	if !validateClientCredentials(clientID, clientSecret) {
		http.Error(w, "Invalid client credentials", http.StatusUnauthorized)
		return
	}

	token, err := generateJWT(clientID)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	// Log token issuance
	auditLog(fmt.Sprintf("Token issued to client_id: %s at %s", clientID, time.Now().Format(time.RFC3339)))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"access_token": "%s", "token_type": "Bearer", "expires_in": 3600}`, token)
}

// Function to handle HTTP token authentication
func tokenAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized: No token provided", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer"))
		claims := &Claims{}

		// Parse the token
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// Ensure the signing method is HMAC (HS256)
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(authTokenSecret), nil
		})

		// Improved error handling for token parsing
		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				http.Error(w, "Unauthorized: Token expired", http.StatusUnauthorized)
			} else if errors.Is(err, jwt.ErrTokenMalformed) || errors.Is(err, jwt.ErrTokenSignatureInvalid) {
				http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
			} else {
				http.Error(w, "Unauthorized: Token parsing error", http.StatusUnauthorized)
			}
			return
		}

		// Check if the token is valid
		if !token.Valid {
			http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
			return
		}

		// Log access to a protected route
		auditLog(fmt.Sprintf("Route accessed: %s by client_id: %s at %s", r.URL.Path, claims.ClientID, time.Now().Format(time.RFC3339)))

		next.ServeHTTP(w, r)
	})
}

// Define a structure for the JSON output
type RoutesResponse struct {
	Routes []string `json:"routes"`
}

// showRoutesCmd returns a Cobra command that lists all registered routes
func showRoutesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show-routes",
		Short: "Show all registered HTTP routes",
		Run: func(cmd *cobra.Command, args []string) {
			initConfig()
			initAuditLogger()
			initURLHandlers()
			if routes == nil {
				fmt.Println("No routes available. Please initialize the server first.")
				return
			}

			if len(routes) == 0 {
				fmt.Println("No routes registered.")
				return
			}
			sort.Strings(routes)
			routes = removeDuplicatesFromSlice(routes)

			// Create the response object
			response := RoutesResponse{
				Routes: routes,
			}

			// Marshal the response object into JSON
			jsonData, err := json.MarshalIndent(response, "", "  ")
			if err != nil {
				fmt.Printf("Error marshaling JSON: %v\n", err)
				return
			}

			// Print the JSON output
			fmt.Println(string(jsonData))

		},
	}
}

// Start the web server with configurable port
func initURLHandlers() {

	unprotectedMux := NewRouteTrackingMux()
	unprotectedMux.HandleFunc("/token", tokenHandler)

	mux := NewRouteTrackingMux()
	mux.HandleFunc("/", checkStatus)
	mux.HandleFunc("/check", checkStatus)
	mux.HandleFunc("/check/disks", checkDisks)
	mux.HandleFunc("/check/dns", checkDNS)

	secureMux := tokenAuthMiddleware(mux)

	// Combine both muxes into a single handler
	mainMux = NewRouteTrackingMux()
	mainMux.Handle("/token", unprotectedMux)
	mainMux.Handle("/", secureMux) // All other routes go through the secure mux

	routes = append(mainMux.Routes(), unprotectedMux.Routes()...)
	routes = append(routes, mux.Routes()...)
	if len(routes) == 0 {
		fmt.Println("No routes registered.")
		return
	}

}

func runServer(port int) {
	//auditLog(fmt.Sprintf("Starting server on port %d...\n", port))
	log.Printf("Starting server on port %d...\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), mainMux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
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

// createClientCmd returns a Cobra command for creating a new client
func createClientCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create-client",
		Short: "Create a new client_id and client_secret and append them to the config file",
		Run: func(cmd *cobra.Command, args []string) {
			initConfig()
			initAuditLogger()
			clientID, clientSecret := generateClientCredentials()

			// Append the new client credentials to the configuration
			appendClientCredentials(clientID, clientSecret)
		},
	}
}

// generateClientCredentials generates a random client_id and client_secret
func generateClientCredentials() (string, string) {
	clientID := generateRandomString(16)
	clientSecret := generateRandomString(32)
	return clientID, clientSecret
}

// generateRandomString generates a random string of specified length
func generateRandomString(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		log.Fatalf("Failed to generate random string: %v", err)
	}
	return hex.EncodeToString(bytes)[:length]
}

// appendClientCredentials appends the new client credentials to the configuration file
func appendClientCredentials(clientID, clientSecret string) {
	// Retrieve the existing clients slice from the configuration
	var clients []map[string]string

	if err := viper.UnmarshalKey("clients", &clients); err != nil {
		log.Fatalf("Error reading clients from config: %s", err)
	}

	// Create a new client entry
	newClient := map[string]string{
		"client_id":     clientID,
		"client_secret": clientSecret,
	}

	// Append the new client to the existing list
	clients = append(clients, newClient)

	// Update the "clients" key in Viper with the new clients list
	viper.Set("clients", clients)

	// Write the updated configuration back to the file
	if err := viper.WriteConfig(); err != nil {
		log.Fatalf("Error writing to config file: %s", err)
	}

	fmt.Printf("New client_id and client_secret added:\nclient_id: %s\nclient_secret: %s\n", clientID, clientSecret)

	// Log the new client creation
	auditLog(fmt.Sprintf("New client created: client_id: %s at %s", clientID, time.Now().Format(time.RFC3339)))
}

// Function to initialize the configuration for Viper
func initConfig() {
	viper.SetConfigName("snh-config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	// Add /usr/local/etc as an additional search path
	viper.AddConfigPath("/usr/local/etc")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	// Load client credentials
	if err := viper.UnmarshalKey("clients", &clients); err != nil {
		log.Fatalf("Error parsing client configuration: %v", err)
	}

	// Load token secret
	authTokenSecret = viper.GetString("authTokenSecret")
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
			initConfig()
			initAuditLogger()
			initURLHandlers()
			runServer(port)
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

	// Add the command to generate and append a new client
	rootCmd.AddCommand(createClientCmd())

	// Add the command to show all registered routes
	rootCmd.AddCommand(showRoutesCmd())

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
		log.Fatalf("Error executing root command: %v", err)
		os.Exit(1)
	}

}
