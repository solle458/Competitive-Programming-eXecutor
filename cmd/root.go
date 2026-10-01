package cmd

import (
	"fmt"
	"os"

	"Competitive-Programming-eXecutor/internal/app"
	"Competitive-Programming-eXecutor/internal/config"

	"github.com/spf13/cobra"
)

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
			cfg, err := config.FindAndLoad()
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
	root.AddCommand(newLibCmd(app))
	return root
}

func Execute(app *app.App) {
	err := rootCmd(app).Execute()
	if err != nil {
		os.Exit(1)
	}
}
