package cli

import (
	"fmt"
	"os"

	"github.com/nhv96/goas/internal/app"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile   string
	think     bool
	stream    bool
	modelName string
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "goas",
	Short: "goas is a CLI tool to talk to AI model.",
	Long:  `A tool that turns your AI model into powerful assistant.`,
	Run: func(cmd *cobra.Command, args []string) {
		a, err := app.NewApplication(&app.Config{
			ModelName: modelName,
			Think:     think,
			Stream:    stream,
		})
		if err != nil {
			panic(err)
		}

		a.Start()
	},
}

// Execute is the program entry point
func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().BoolVarP(&think, "think", "t", true, "To think or not to think")
	RootCmd.PersistentFlags().BoolVarP(&stream, "stream", "s", true, "To stream or not to stream")
	RootCmd.PersistentFlags().StringVarP(&modelName, "model", "m", "gemma4:e2b", "The model name to chat with")
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".goas" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".goas")
	}

	viper.AutomaticEnv() // Read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
