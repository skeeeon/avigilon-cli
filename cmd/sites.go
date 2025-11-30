package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"avigilon-cli/internal/client"
)

var sitesCmd = &cobra.Command{
	Use:   "sites",
	Short: "List all ACC Sites (Clusters)",
	Run: func(cmd *cobra.Command, args []string) {
		baseUrl := viper.GetString("base_url")
		session := viper.GetString("session_id")

		if baseUrl == "" || session == "" {
			fmt.Println("Error: Not logged in. Please run 'avigilon-cli login' first.")
			os.Exit(1)
		}

		api := client.New(client.ClientConfig{BaseURL: baseUrl})
		api.HTTP.SetHeader("x-avg-session", session)

		sites, err := api.GetSites()
		if err != nil {
			fmt.Printf("Error fetching sites: %v\n", err)
			os.Exit(1)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME")
		fmt.Fprintln(w, "--\t----")

		for _, site := range sites {
			fmt.Fprintf(w, "%s\t%s\n", site.ID, site.Name)
		}
		w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(sitesCmd)
}
