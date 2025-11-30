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
	cameraID       string
	outputFile     string
	recordIDs      string
	recordDuration int
	recordStop     bool
)

// Helper to initialize client and retrieve session
// Returns the client instance and the raw session string (needed for some payloads)
func setupCameraClient() (*client.AvigilonClient, string) {
	baseUrl := viper.GetString("base_url")
	session := viper.GetString("session_id")

	if baseUrl == "" || session == "" {
		fmt.Println("Error: Not logged in. Please run 'avigilon-cli login' first.")
		os.Exit(1)
	}

	api := client.New(client.ClientConfig{BaseURL: baseUrl})
	// Inject session header for standard requests
	api.HTTP.SetHeader("x-avg-session", session)

	return api, session
}

// Parent Command
var camerasCmd = &cobra.Command{
	Use:   "cameras",
	Short: "Manage cameras",
	Long:  `List cameras, take snapshots, or trigger manual recordings.`,
}

// List Command
var camerasListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all cameras",
	Run: func(cmd *cobra.Command, args []string) {
		api, _ := setupCameraClient()

		cameras, err := api.GetCameras()
		if err != nil {
			fmt.Printf("Error fetching cameras: %v\n", err)
			os.Exit(1)
		}

		// --- JSON OUTPUT ---
		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			if err := enc.Encode(cameras); err != nil {
				fmt.Printf("Error encoding JSON: %v\n", err)
				os.Exit(1)
			}
			return
		}
		// -------------------

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tMODEL\tSTATUS\tIP")
		fmt.Fprintln(w, "--\t----\t-----\t------\t--")

		for _, cam := range cameras {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				cam.ID,
				cam.Name,
				cam.Model,
				cam.ConnectionState,
				cam.IPAddress,
			)
		}
		w.Flush()
	},
}

// Snapshot Command
var camerasSnapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Take a JPEG snapshot from a camera",
	Example: `  avigilon-cli cameras snapshot --id "camera_id_string" --output "image.jpg"`,
	Run: func(cmd *cobra.Command, args []string) {
		api, _ := setupCameraClient()

		fmt.Printf("Requesting snapshot for Camera ID: %s ...\n", cameraID)

		imgData, err := api.GetSnapshot(cameraID)
		if err != nil {
			fmt.Printf("Error getting snapshot: %v\n", err)
			os.Exit(1)
		}

		if err := os.WriteFile(outputFile, imgData, 0644); err != nil {
			fmt.Printf("Error writing file: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Snapshot saved to %s\n", outputFile)
	},
}

// Record Command
var camerasRecordCmd = &cobra.Command{
	Use:   "record",
	Short: "Trigger manual recording on cameras",
	Long:  `Start or stop manual recording on one or more cameras.`,
	Example: `  avigilon-cli cameras record --ids "id1,id2" --seconds 60
  avigilon-cli cameras record --ids "id1" --stop`,
	Run: func(cmd *cobra.Command, args []string) {
		api, session := setupCameraClient()

		// Parse IDs from comma-separated string
		ids := strings.Split(recordIDs, ",")
		// Clean whitespace
		var cleanIDs []string
		for _, id := range ids {
			trimmed := strings.TrimSpace(id)
			if trimmed != "" {
				cleanIDs = append(cleanIDs, trimmed)
			}
		}

		if len(cleanIDs) == 0 {
			fmt.Println("Error: No valid Camera IDs provided.")
			os.Exit(1)
		}

		action := "START"
		if recordStop {
			action = "STOP"
		}

		if action == "START" {
			fmt.Printf("Triggering recording for %d seconds on cameras: %v\n", recordDuration, cleanIDs)
		} else {
			fmt.Printf("Stopping manual recording on cameras: %v\n", cleanIDs)
		}

		// Call Client
		err := api.TriggerManualRecording(session, cleanIDs, action, recordDuration)
		if err != nil {
			fmt.Printf("Error triggering recording: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Success.")
	},
}

func init() {
	// Register Parent
	rootCmd.AddCommand(camerasCmd)

	// Register Subcommands
	camerasCmd.AddCommand(camerasListCmd)
	camerasCmd.AddCommand(camerasSnapshotCmd)
	camerasCmd.AddCommand(camerasRecordCmd)

	// Flags for Snapshot
	camerasSnapshotCmd.Flags().StringVar(&cameraID, "id", "", "ID of the camera")
	camerasSnapshotCmd.Flags().StringVar(&outputFile, "output", "snapshot.jpg", "Output filename")
	_ = camerasSnapshotCmd.MarkFlagRequired("id")

	// Flags for Record
	camerasRecordCmd.Flags().StringVar(&recordIDs, "ids", "", "Comma separated list of Camera IDs")
	camerasRecordCmd.Flags().IntVar(&recordDuration, "seconds", 300, "Duration in seconds (default 5 mins)")
	camerasRecordCmd.Flags().BoolVar(&recordStop, "stop", false, "Stop recording instead of starting")
	_ = camerasRecordCmd.MarkFlagRequired("ids")
}
