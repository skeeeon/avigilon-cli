package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"avigilon-cli/internal/config"
)

var cfgFile string
var jsonOutput bool 

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "avigilon-cli",
	Short: "A CLI for interacting with Avigilon Web Endpoint API",
	Long: `Manage cameras, alarms, and users on your Avigilon Control Center 
via the Web Endpoint Service.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(func() { config.InitConfig(cfgFile) })
	
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.avigilon-cli.yaml)")
	
	// Add the persistent flag here
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output results as JSON")
}
