package cmd

import (
	"log"

	"github.com/spf13/viper"
)

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
