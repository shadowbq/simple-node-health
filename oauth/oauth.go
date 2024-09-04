// package: oauth
// file: oauth.go
// Description: This file contains the implementation of the OAuth 2.0 Authorization Server.
package oauth

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/shadowbq/simple-node-health/audit"
	"github.com/spf13/viper"
)

type Client struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
}

type Claims struct {
	ClientID string `json:"client_id"`
	jwt.RegisteredClaims
}

var Clients []Client

var authTokenSecret string

// Function GenerateJWT creates a new JWT token
func generateJWT(clientID string, authTokenSecret string) (string, error) {
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

	if err := viper.UnmarshalKey("clients", &Clients); err != nil {
		log.Fatalf("Error reading clients from config: %s", err)
	}
	// Log out the clients size
	log.Printf(fmt.Sprintf("Clients size: %d", len(Clients))) // Verbose logging

	for _, client := range Clients {
		if client.ClientID == clientID && client.ClientSecret == clientSecret {
			return true
		}
	}
	return false
}

// Function to handle Token and call generator for JWT token
func TokenHandler(w http.ResponseWriter, r *http.Request) {
	clientID := r.FormValue("client_id")
	clientSecret := r.FormValue("client_secret")
	log.Printf(fmt.Sprintf("Route requested: %s by client_id: %s:%s at %s", r.URL.Path, clientID, clientSecret, time.Now().Format(time.RFC3339))) // Verbose logging

	if !validateClientCredentials(clientID, clientSecret) {
		log.Printf(fmt.Sprintf("Invalid client credentials: %s:%s", clientID, clientSecret)) // Verbose logging
		http.Error(w, "Invalid client credentials", http.StatusUnauthorized)
		return
	}

	token, err := generateJWT(clientID, authTokenSecret)
	if err != nil {
		log.Printf(fmt.Sprintf("Error generating token: %v", err)) // Verbose logging
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	// Log token issuance
	audit.AuditLog(fmt.Sprintf("Token issued to client_id: %s at %s", clientID, time.Now().Format(time.RFC3339)))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"access_token": "%s", "token_type": "Bearer", "expires_in": 3600}`, token)
}

// Function to handle HTTP token authentication
func TokenAuthMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized: No token provided", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer"))
		claims := &Claims{}

		//audit.AuditLog(fmt.Sprintf("Route accessed: %s by client_id: %s at %s", r.URL.Path, claims.ClientID, time.Now().Format(time.RFC3339)))

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
				audit.AuditLog(fmt.Sprintf("Unauthorized: Token expired: %s", err))
				http.Error(w, "Unauthorized: Token expired", http.StatusUnauthorized)
			} else if errors.Is(err, jwt.ErrTokenMalformed) || errors.Is(err, jwt.ErrTokenSignatureInvalid) {
				audit.AuditLog(fmt.Sprintf("Unauthorized: Invalid token: %s", err))
				http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
			} else {
				audit.AuditLog(fmt.Sprintf("Unauthorized: Token parsing error: %s", err))
				http.Error(w, "Unauthorized: Token parsing error", http.StatusUnauthorized)
			}
			return
		}

		// Check if the token is valid
		if !token.Valid {
			audit.AuditLog("Unauthorized: Invalid token")
			http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
			return
		}

		// Log access to a protected route
		audit.AuditLog(fmt.Sprintf("Route accessed: %s by client_id: %s at %s", r.URL.Path, claims.ClientID, time.Now().Format(time.RFC3339)))

		next.ServeHTTP(w, r)
	})
}
