// Package oauth - client.go - The code in this file is used to create a new client_id and client_secret and append them to the configuration file.
package oauth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/shadowbq/simple-node-health/audit"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// CreateClientInternalCmd returns a Cobra command for creating a new client
func CreateClientInternalCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create-client-internal",
		Short: "Create a new client_id and client_secret and append them to the config file",
		Run: func(cmd *cobra.Command, args []string) {
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
	audit.AuditLog(fmt.Sprintf("New client created: client_id: %s at %s", clientID, time.Now().Format(time.RFC3339)))
}
