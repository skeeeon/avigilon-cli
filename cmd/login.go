package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"avigilon-cli/internal/client"
	"avigilon-cli/internal/config"
)

// Variables to hold flag values
var (
	host   string
	user   string
	pass   string
	nonce  string
	key    string
	intID  string
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with the Avigilon Server",
	Long: `Authenticates using the provided credentials, generates the required 
cryptographic signature, and saves the session token locally for future commands.

Example:
  avigilon-cli login --host "https://10.0.0.5/mt/api/rest/v1" --username admin --password pass --nonce myNonce --key myKey`,
	Run: func(cmd *cobra.Command, args []string) {
		// Clean up input host (remove trailing slash if present)
		host = strings.TrimRight(host, "/")

		// 1. Construct the configuration object from flags
		cfg := client.ClientConfig{
			BaseURL:       host,
			Username:      user,
			Password:      pass,
			UserNonce:     nonce,
			UserKey:       key,
			IntegrationID: intID,
		}

		fmt.Printf("Authenticating against %s as user '%s'...\n", host, user)

		// 2. Initialize Client
		api := client.New(cfg)

		// 3. Perform Login
		// Note: This relies on internal/client/client.go Login() being updated to return (string, error)
		sessionID, err := api.Login()
		if err != nil {
			log.Fatalf("Fatal: Login failed: %v", err)
		}

		fmt.Println("Login successful. Saving configuration...")

		// 4. Update Viper Configuration
		// We save the Base URL so subsequent commands (like 'cameras') know where to connect.
		viper.Set("base_url", host)

		// 5. Persist Session and Config to file
		// We use the helper from internal/config to handle file creation/writing
		if err := config.SaveSession(sessionID); err != nil {
			log.Fatalf("Failed to save configuration file: %v", err)
		}

		fmt.Printf("Session saved. You can now run commands like './avigilon-cli cameras'.\n")
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)

	// Define Flags
	// We use local flags because these are specific only to the login action.
	loginCmd.Flags().StringVar(&host, "host", "", "API Base URL (e.g. https://192.168.1.50/mt/api/rest/v1)")
	loginCmd.Flags().StringVarP(&user, "username", "u", "administrator", "ACC Username")
	loginCmd.Flags().StringVarP(&pass, "password", "p", "", "ACC Password")
	loginCmd.Flags().StringVar(&nonce, "nonce", "", "User Nonce (from Avigilon Integrator Config)")
	loginCmd.Flags().StringVar(&key, "key", "", "User Key (from Avigilon Integrator Config)")
	loginCmd.Flags().StringVar(&intID, "integration-id", "", "Integration ID (optional, leave empty if not used)")

	// Mark required flags to ensure the user provides necessary info
	_ = loginCmd.MarkFlagRequired("host")
	_ = loginCmd.MarkFlagRequired("password")
	_ = loginCmd.MarkFlagRequired("nonce")
	_ = loginCmd.MarkFlagRequired("key")
}
