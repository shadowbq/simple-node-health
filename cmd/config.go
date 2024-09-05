package cmd

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

var (
	authTokenSecret   string
	ClientsFromConfig []Client
)

type Client struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
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

	// Load client credentials into the Clients slice
	if err := viper.UnmarshalKey("clients", &ClientsFromConfig); err != nil {
		log.Fatalf("Error parsing client configuration: %v", err)
	}

	// log the clients size
	log.Printf("Clients size from Config: %d", len(ClientsFromConfig))

	if len(ClientsFromConfig) == 0 {
		log.Printf("No clients found. ")

		// Check if insecure mode is enabled in the config
		if viper.GetBool("insecure") {
			log.Println("Insecure mode enabled. No client credentials required.")
		} else {
			log.Println("Either enable insecure mode, or run: 'simple-node-health create-client'")
			os.Exit(1)
		}
	} // Verbose logging

	// Load token secret
	authTokenSecret = viper.GetString("authTokenSecret")
}
