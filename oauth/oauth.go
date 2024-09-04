package oauth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

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
	commonutils.auditLog(fmt.Sprintf("Token issued to client_id: %s at %s", clientID, time.Now().Format(time.RFC3339)))

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
		commonutils.auditLog(fmt.Sprintf("Route accessed: %s by client_id: %s at %s", r.URL.Path, claims.ClientID, time.Now().Format(time.RFC3339)))

		next.ServeHTTP(w, r)
	})
}
