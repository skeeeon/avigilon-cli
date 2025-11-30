package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"avigilon-cli/internal/client"
)

var serversCmd = &cobra.Command{
	Use:   "servers",
	Short: "List all Servers in the cluster",
	Run: func(cmd *cobra.Command, args []string) {
		baseUrl := viper.GetString("base_url")
		session := viper.GetString("session_id")

		if baseUrl == "" || session == "" {
			fmt.Println("Error: Not logged in. Please run 'avigilon-cli login' first.")
			os.Exit(1)
		}

		api := client.New(client.ClientConfig{BaseURL: baseUrl})
		api.HTTP.SetHeader("x-avg-session", session)

		servers, err := api.GetServers()
		if err != nil {
			fmt.Printf("Error fetching servers: %v\n", err)
			os.Exit(1)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME")
		fmt.Fprintln(w, "--\t----")

		for _, srv := range servers {
			fmt.Fprintf(w, "%s\t%s\n", srv.ID, srv.Name)
		}
		w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(serversCmd)
}
