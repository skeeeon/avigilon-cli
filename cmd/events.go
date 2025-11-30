package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"avigilon-cli/internal/client"
	"avigilon-cli/pkg/models"
)

var (
	eventSince  string
	eventTopics string
)

var eventsCmd = &cobra.Command{
	Use:   "events",
	Short: "Search historical events",
	Long:  `Search for events (motion, system, errors) over a specific time range across all servers.`,
}

var eventsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List events from history",
	Run: func(cmd *cobra.Command, args []string) {
		baseUrl := viper.GetString("base_url")
		session := viper.GetString("session_id")

		if baseUrl == "" || session == "" {
			fmt.Println("Error: Not logged in. Please run 'avigilon-cli login' first.")
			os.Exit(1)
		}

		api := client.New(client.ClientConfig{BaseURL: baseUrl})
		api.HTTP.SetHeader("x-avg-session", session)

		// 1. Get Servers
		servers, err := api.GetServers()
		if err != nil {
			fmt.Printf("Error discovering servers: %v\n", err)
			os.Exit(1)
		}

		// 2. Setup Time Range
		duration, err := time.ParseDuration(eventSince)
		if err != nil {
			fmt.Printf("Error parsing duration: %v\n", err)
			os.Exit(1)
		}
		to := time.Now().UTC()
		from := to.Add(-duration)

		// 3. Parse Topics (Clean spaces)
		var topicsSlice []string
		if eventTopics != "" {
			rawSlice := strings.Split(eventTopics, ",")
			for _, t := range rawSlice {
				trimmed := strings.TrimSpace(t)
				if trimmed != "" {
					topicsSlice = append(topicsSlice, trimmed)
				}
			}
		}

		fmt.Printf("Searching %d servers from %s to %s (UTC)...\n", len(servers), from.Format("15:04"), to.Format("15:04"))

		// 4. Aggregate Events
		var allEvents []models.Event
		for _, srv := range servers {
			evts, err := api.GetEvents(srv.ID, from, to, topicsSlice)
			if err != nil {
				fmt.Printf("Warning: Failed to query server %s: %v\n", srv.Name, err)
				continue
			}
			allEvents = append(allEvents, evts...)
		}

		// --- JSON OUTPUT ---
		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			if err := enc.Encode(allEvents); err != nil {
				fmt.Printf("Error encoding JSON: %v\n", err)
				os.Exit(1)
			}
			return
		}

		if len(allEvents) == 0 {
			fmt.Println("No events found in this time range.")
			return
		}

		// 5. Print Table
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "TIMESTAMP\tTYPE\tSOURCE\tSERVER")
		fmt.Fprintln(w, "---------\t----\t------\t------")

		for _, e := range allEvents {
			ts := e.Timestamp
			// Parse ISO8601 back to local time for display
			if t, err := time.Parse(time.RFC3339, e.Timestamp); err == nil {
				// UPDATED: Format now includes Date + Time (YYYY-MM-DD HH:MM:SS)
				ts = t.Local().Format("2006-01-02 15:04:05")
			}

			source := e.CameraID
			if source == "" {
				source = e.UserName
			}
			if source == "" {
				source = "System"
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", ts, e.Type, source, e.Server)
		}
		w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(eventsCmd)
	eventsCmd.AddCommand(eventsListCmd)

	eventsListCmd.Flags().StringVar(&eventSince, "since", "1h", "Look back duration (e.g. 30m, 1h, 24h)")
	eventsListCmd.Flags().StringVar(&eventTopics, "topics", "", "Comma separated list of event topics")
}
