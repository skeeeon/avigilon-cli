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

// Variables for snapshot command
var (
	cameraID   string
	outputFile string
)

// Parent Command
var camerasCmd = &cobra.Command{
	Use:   "cameras",
	Short: "Manage cameras",
	Long:  `List cameras or take snapshots.`,
}

// List Command (Moved logic from original camerasCmd to here)
var camerasListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all cameras",
	Run: func(cmd *cobra.Command, args []string) {
		baseUrl := viper.GetString("base_url")
		session := viper.GetString("session_id")

		if baseUrl == "" || session == "" {
			fmt.Println("Error: Not logged in. Please run 'avigilon-cli login' first.")
			os.Exit(1)
		}

		api := client.New(client.ClientConfig{BaseURL: baseUrl})
		api.HTTP.SetHeader("x-avg-session", session)

		cameras, err := api.GetCameras()
		if err != nil {
			fmt.Printf("Error fetching cameras: %v\n", err)
			os.Exit(1)
		}

		// --- JSON OUTPUT LOGIC ---
		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			if err := enc.Encode(cameras); err != nil {
				fmt.Printf("Error encoding JSON: %v\n", err)
				os.Exit(1)
			}
			return // Exit here so we don't print the table
		}
		// -------------------------

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
		baseUrl := viper.GetString("base_url")
		session := viper.GetString("session_id")

		if baseUrl == "" || session == "" {
			fmt.Println("Error: Not logged in. Please run 'avigilon-cli login' first.")
			os.Exit(1)
		}

		api := client.New(client.ClientConfig{BaseURL: baseUrl})
		api.HTTP.SetHeader("x-avg-session", session)

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

func init() {
	rootCmd.AddCommand(camerasCmd)
	
	// Add subcommands
	camerasCmd.AddCommand(camerasListCmd)
	camerasCmd.AddCommand(camerasSnapshotCmd)

	// Flags for Snapshot
	camerasSnapshotCmd.Flags().StringVar(&cameraID, "id", "", "ID of the camera")
	camerasSnapshotCmd.Flags().StringVar(&outputFile, "output", "snapshot.jpg", "Output filename")
	_ = camerasSnapshotCmd.MarkFlagRequired("id")
}
