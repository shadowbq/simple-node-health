package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"

	"github.com/shadowbq/simple-node-health/audit"
	"github.com/shadowbq/simple-node-health/helpers"
	"github.com/shadowbq/simple-node-health/oauth"
	"github.com/shadowbq/simple-node-health/parsers"
	"github.com/spf13/cobra"
)

var (
	routes  []string
	mux     *RouteTrackingMux
	mainMux *RouteTrackingMux
)

//var secureMux *http.ServeMux

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
			audit.InitAuditLogger()
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
			routes = helpers.RemoveDuplicatesFromSlice(routes)

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
	unprotectedMux.HandleFunc("/token", oauth.TokenHandler)

	mux := NewRouteTrackingMux()
	mux.HandleFunc("/", parsers.HTTPCheckStatus)
	mux.HandleFunc("/check", parsers.HTTPCheckStatus)
	mux.HandleFunc("/check/disks", parsers.HTTPCheckDisks)
	mux.HandleFunc("/check/dns", parsers.HTTPCheckDNS)

	secureMux := oauth.TokenAuthMiddleware(mux)

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
	audit.AuditLog(fmt.Sprintf("Starting server on port %d...\n", port))
	log.Printf("Starting server on port %d...\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), mainMux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
		os.Exit(1)
	}
}
