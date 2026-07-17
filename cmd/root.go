/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"Competitive-Programming-eXecutor/internal/app"
	"Competitive-Programming-eXecutor/internal/config"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
func rootCmd(app *app.App) *cobra.Command {
	root := &cobra.Command{
		Use:   "cpx",
		Short: "Competitive Programming eXecutor",
		Long: `cpx helps you set up AtCoder contests, run sample tests,
merge libraries, and submit (or copy) solutions.`,
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Name() == "init" {
				return nil
			}
			cfg, err := config.FindAndLoad() // cwd から親へ .config.yaml を探す
			if err != nil {
				return fmt.Errorf("not initialized: run `cpx init` first")
			}
			app.Config = cfg
			return nil
		},
	}
	root.AddCommand(initCmd(app))
	root.AddCommand(mergeCmd(app))
	root.AddCommand(setupCmd(app))
	root.AddCommand(testCmd(app))
	root.AddCommand(submitCmd(app))
	return root
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(app *app.App) {
	err := rootCmd(app).Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.Competitive-Programming-eXecutor.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
