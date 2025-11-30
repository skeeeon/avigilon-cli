package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"avigilon-cli/internal/client"
)

// Variables to hold flag values
var (
	alarmID     string
	alarmAction string
	alarmNote   string
)

// Helper to get authenticated client using stored config
func getAlarmClient() *client.AvigilonClient {
	baseUrl := viper.GetString("base_url")
	session := viper.GetString("session_id")

	if baseUrl == "" || session == "" {
		fmt.Println("Error: Not logged in. Please run 'avigilon-cli login' first.")
		os.Exit(1)
	}

	api := client.New(client.ClientConfig{BaseURL: baseUrl})
	api.HTTP.SetHeader("x-avg-session", session)
	return api
}

// Parent Command
var alarmsCmd = &cobra.Command{
	Use:   "alarms",
	Short: "Manage Alarms",
	Long:  `List active alarms or perform actions (acknowledge/purge) on them.`,
}

// List Command
var alarmsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List active alarms",
	Run: func(cmd *cobra.Command, args []string) {
		api := getAlarmClient()

		alarms, err := api.GetAlarms()
		if err != nil {
			fmt.Printf("Error fetching alarms: %v\n", err)
			os.Exit(1)
		}

		// --- JSON OUTPUT ---
		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			if err := enc.Encode(alarms); err != nil {
				fmt.Printf("Error encoding JSON: %v\n", err)
				os.Exit(1)
			}
			return
		}
		// -------------------

		if len(alarms) == 0 {
			fmt.Println("No active alarms.")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tSTATE\tTRIGGER TIME")
		fmt.Fprintln(w, "--\t----\t-----\t------------")

		for _, a := range alarms {
			name := a.Name
			if name == "" {
				name = "[No Name]"
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				a.ID,
				name,
				a.State,
				a.TriggerTime,
			)
		}
		w.Flush()
	},
}

// Update Command
var alarmsUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Perform action on an alarm",
	Example: `  avigilon-cli alarms update --id "zFgy_123" --action "ACKNOWLEDGE" --note "Reviewing"`,
	Run: func(cmd *cobra.Command, args []string) {
		session := viper.GetString("session_id")
		api := getAlarmClient()

		fmt.Printf("Sending action '%s' to Alarm %s...\n", alarmAction, alarmID)

		err := api.UpdateAlarm(session, alarmID, alarmAction, alarmNote)
		if err != nil {
			fmt.Printf("Error updating alarm: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Alarm updated successfully.")
	},
}

func init() {
	// Register Parent
	rootCmd.AddCommand(alarmsCmd)

	// Register List
	alarmsCmd.AddCommand(alarmsListCmd)

	// Register Update
	alarmsCmd.AddCommand(alarmsUpdateCmd)
	alarmsUpdateCmd.Flags().StringVar(&alarmID, "id", "", "Alarm ID to update")
	alarmsUpdateCmd.Flags().StringVar(&alarmAction, "action", "ACKNOWLEDGE", "Action to perform (ACKNOWLEDGE, PURGE, DISMISS)")
	alarmsUpdateCmd.Flags().StringVar(&alarmNote, "note", "", "Optional note/comment")
	_ = alarmsUpdateCmd.MarkFlagRequired("id")
}
