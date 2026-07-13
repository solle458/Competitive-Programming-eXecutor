/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"Competitive-Programming-eXecutor/internal/app"
	"Competitive-Programming-eXecutor/internal/test"
	"errors"

	"github.com/spf13/cobra"
)

// testCmd represents the test command
func testCmd(app *app.App) *cobra.Command {
	var (
		lang      string
		timeLimit int
	)
	var (
		problemIDRequired = errors.New("problem id is required")
	)

	cmd := &cobra.Command{
		Use:   "test",
		Short: "Test the competitive programming",
		Long:  `Test the competitive programming`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return problemIDRequired
			}
			problemID := args[0]
			executableFilePath, err := test.Compile(problemID, lang, app.Config)
			if err != nil {
				return err
			}
			executionTimes, err := test.Run(problemID, executableFilePath, lang)
			if err != nil {
				return err
			}
			err = test.Compare(problemID, executionTimes, timeLimit)
			if err != nil {
				return err
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&lang, "lang", "l", "cpp", "language of the source code")
	cmd.Flags().IntVarP(&timeLimit, "time-limit", "t", 2, "time limit in seconds")
	return cmd
}
