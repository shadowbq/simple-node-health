package cmd

import (
	"fmt"
	"os"

	"github.com/shadowbq/simple-node-health/audit"
	"github.com/shadowbq/simple-node-health/oauth"
	"github.com/shadowbq/simple-node-health/parsers"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	port    int
	domain  string
	verbose bool
)

// Root command
var rootCmd = &cobra.Command{
	Use:   "simple-node-health",
	Short: "A simple tool to check hardware EXT4 devices and run DNS queries",
	Run: func(cmd *cobra.Command, args []string) {
		initConfig()
		audit.InitAuditLogger()
		initURLHandlers()
		runServer(port)
	},
}

var CreateClientCmd = &cobra.Command{
	Use:   "create-client",
	Short: "Create a new client_id and client_secret and append them to the config file",
	Run: func(cmd *cobra.Command, args []string) {
		createClient := oauth.CreateClientInternalCmd()
		initConfig()
		audit.InitAuditLogger()
		createClient.Run(createClient, args)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {

	rootCmd.AddCommand(CreateClientCmd)

	// The Create Client Comand needs to be init so CreateClientInternalCmd is wrapped by CreateClientCmd
	// rootCmd.AddCommand(oauth.CreateClientInternalCmd())

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
		Run:   parsers.CmdCheckStatus,
	}

	// Subcommand: checkdisks
	var checkDisksCmd = &cobra.Command{
		Use:   "disks",
		Short: "Check EXT4 devices for read-only mode",
		Run:   parsers.CmdCheckDisks,
	}

	// Subcommand: checkdns
	var checkDNSCmd = &cobra.Command{
		Use:   "dns",
		Short: "Run a DNS query for the specified domain",
		Run:   parsers.CmdCheckDNS,
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

}
