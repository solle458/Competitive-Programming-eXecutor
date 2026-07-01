/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"Competitive-Programming-eXecutor/internal/app"

	"github.com/spf13/cobra"
)

// testCmd represents the test command
func testCmd(app *app.App) *cobra.Command {
	var lang string
	var onlyCompile bool
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Test the competitive programming",
		Long:  `Test the competitive programming`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	cmd.Flags().StringVarP(&lang, "lang", "l", "cpp", "language of the source code")
	cmd.Flags().BoolVarP(&onlyCompile, "only-compile", "c", false, "only compile the source code")
	return cmd
}
