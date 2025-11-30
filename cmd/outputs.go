package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"avigilon-cli/internal/client"
)

var (
	outputTargetID string
	outputIsCamera bool
)

// Parent Command
var outputsCmd = &cobra.Command{
	Use:   "outputs",
	Short: "Manage digital outputs",
	Long:  `Trigger digital outputs connected to cameras or I/O modules.`,
}

// Trigger Command
var outputsTriggerCmd = &cobra.Command{
	Use:   "trigger",
	Short: "Trigger a digital output",
	Example: `  avigilon-cli outputs trigger --id "camera_id_here" --camera
  avigilon-cli outputs trigger --id "specific_output_entity_id"`,
	Run: func(cmd *cobra.Command, args []string) {
		baseUrl := viper.GetString("base_url")
		session := viper.GetString("session_id")

		if baseUrl == "" || session == "" {
			fmt.Println("Error: Not logged in. Please run 'avigilon-cli login' first.")
			os.Exit(1)
		}

		api := client.New(client.ClientConfig{BaseURL: baseUrl})
		api.HTTP.SetHeader("x-avg-session", session)

		targetType := "Digital Output Entity"
		if outputIsCamera {
			targetType = "All Outputs on Camera"
		}

		fmt.Printf("Triggering %s (%s)...\n", targetType, outputTargetID)

		err := api.TriggerDigitalOutput(session, outputTargetID, outputIsCamera)
		if err != nil {
			fmt.Printf("Error triggering output: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Output triggered successfully.")
	},
}

func init() {
	rootCmd.AddCommand(outputsCmd)
	outputsCmd.AddCommand(outputsTriggerCmd)

	outputsTriggerCmd.Flags().StringVar(&outputTargetID, "id", "", "ID of the Camera or Digital Output")
	outputsTriggerCmd.Flags().BoolVar(&outputIsCamera, "camera", false, "Set this flag if the ID provided is a Camera ID (triggers all attached outputs)")
	
	_ = outputsTriggerCmd.MarkFlagRequired("id")
}
