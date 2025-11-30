package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"avigilon-cli/internal/client"
)

// Variables to hold flag values
var (
	webhookURL      string
	webhookToken    string
	webhookTopics   string
	webhookID       string
	webhookHBEnable bool
	webhookHBFreq   int
)

// Helper to get authenticated client using stored config
func getClient() *client.AvigilonClient {
	baseUrl := viper.GetString("base_url")
	session := viper.GetString("session_id")

	if baseUrl == "" || session == "" {
		fmt.Println("Error: Not logged in. Please run 'avigilon-cli login' first.")
		os.Exit(1)
	}

	api := client.New(client.ClientConfig{BaseURL: baseUrl})
	// Inject the session header for authentication
	api.HTTP.SetHeader("x-avg-session", session)
	return api
}

// Parent Command
var webhooksCmd = &cobra.Command{
	Use:   "webhooks",
	Short: "Manage event notification webhooks",
	Long:  `List, Create, and Delete webhooks for subscribing to system events.`,
}

// List Command
var webhooksListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registered webhooks",
	Run: func(cmd *cobra.Command, args []string) {
		api := getClient()
		
		hooks, err := api.GetWebhooks()
		if err != nil {
			fmt.Printf("Error fetching webhooks: %v\n", err)
			os.Exit(1)
		}

		// --- JSON OUTPUT ---
		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			if err := enc.Encode(hooks); err != nil {
				fmt.Printf("Error encoding JSON: %v\n", err)
				os.Exit(1)
			}
			return
		}
		// -------------------

		if len(hooks) == 0 {
			fmt.Println("No webhooks found.")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tURL\tTOPICS")
		fmt.Fprintln(w, "--\t---\t------")

		for _, h := range hooks {
			topics := "ALL"
			// Safely access nested struct fields
			if h.EventTopics != nil && len(h.EventTopics.Include) > 0 {
				topics = strings.Join(h.EventTopics.Include, ",")
			}
			
			fmt.Fprintf(w, "%s\t%s\t%s\n", h.ID, h.URL, topics)
		}
		w.Flush()
	},
}

// Create Command
var webhooksCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new webhook",
	Example: `  avigilon-cli webhooks create --url "http://myserver.com/api" --topics "ALL"
  avigilon-cli webhooks create --url "http://myserver.com/api" --heartbeat=false
  avigilon-cli webhooks create --url "http://myserver.com/api" --token "my-custom-secret"`,
	Run: func(cmd *cobra.Command, args []string) {
		// 1. Get Session ID from config (Required for body payload)
		session := viper.GetString("session_id")
		
		// 2. Initialize Client
		api := getClient()
		
		// 3. Process Topics
		topicsSlice := strings.Split(webhookTopics, ",")
		// If empty or just whitespace, default to ALL
		if len(topicsSlice) == 0 || (len(topicsSlice) == 1 && strings.TrimSpace(topicsSlice[0]) == "") {
			topicsSlice = []string{"ALL"}
		}

		// Clean whitespace from topics
		for i := range topicsSlice {
			topicsSlice[i] = strings.TrimSpace(topicsSlice[i])
		}

		fmt.Printf("Creating webhook for URL: %s ...\n", webhookURL)
		fmt.Printf("Configuration: Heartbeat=%t (%dms), Token=%s\n", webhookHBEnable, webhookHBFreq, webhookToken)

		// 4. Call API with all parameters
		err := api.CreateWebhook(session, webhookURL, webhookToken, topicsSlice, webhookHBEnable, webhookHBFreq)
		if err != nil {
			fmt.Printf("Error creating webhook: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Println("Webhook created successfully.")
	},
}

// Delete Command
var webhooksDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a webhook by ID",
	Example: `  avigilon-cli webhooks delete --id "webhook_id_string"`,
	Run: func(cmd *cobra.Command, args []string) {
		api := getClient()
		
		fmt.Printf("Deleting webhook ID: %s ...\n", webhookID)
		
		err := api.DeleteWebhook(webhookID)
		if err != nil {
			fmt.Printf("Error deleting webhook: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Webhook deleted successfully.")
	},
}

func init() {
	// Register parent
	rootCmd.AddCommand(webhooksCmd)
	
	// Register List
	webhooksCmd.AddCommand(webhooksListCmd)
	
	// Register Create
	webhooksCmd.AddCommand(webhooksCreateCmd)
	webhooksCreateCmd.Flags().StringVar(&webhookURL, "url", "", "Target URL for the webhook")
	// Set a default token to satisfy API requirements if user doesn't care
	webhooksCreateCmd.Flags().StringVar(&webhookToken, "token", "avigilon-webhook-token", "Auth token to send to the target (required by API)")
	webhooksCreateCmd.Flags().StringVar(&webhookTopics, "topics", "ALL", "Comma separated list of topics")
	// Heartbeat Flags
	webhooksCreateCmd.Flags().BoolVar(&webhookHBEnable, "heartbeat", true, "Enable heartbeat messages")
	webhooksCreateCmd.Flags().IntVar(&webhookHBFreq, "heartbeat-freq", 3600000, "Heartbeat frequency in milliseconds")
	
	_ = webhooksCreateCmd.MarkFlagRequired("url")

	// Register Delete
	webhooksCmd.AddCommand(webhooksDeleteCmd)
	webhooksDeleteCmd.Flags().StringVar(&webhookID, "id", "", "ID of the webhook to delete")
	_ = webhooksDeleteCmd.MarkFlagRequired("id")
}
