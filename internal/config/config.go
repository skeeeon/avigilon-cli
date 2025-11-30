package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// InitConfig reads in config file and ENV variables if set.
func InitConfig(cfgFile string) {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".avigilon-cli" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".avigilon-cli")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		// Config loaded successfully
	}
}

// SaveSession updates the config file with the new session ID
func SaveSession(sessionID string) error {
	viper.Set("session_id", sessionID)
	
	// Ensure the file exists before writing
	if err := viper.WriteConfig(); err != nil {
		// If file doesn't exist, create it
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return viper.SafeWriteConfig()
		}
		// If it exists but failed to write, try writing to default path
		home, _ := os.UserHomeDir()
		path := filepath.Join(home, ".avigilon-cli.yaml")
		return viper.WriteConfigAs(path)
	}
	return nil
}
