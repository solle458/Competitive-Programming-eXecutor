/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"Competitive-Programming-eXecutor/internal/app"
	"Competitive-Programming-eXecutor/internal/submit"
	"errors"

	"github.com/spf13/cobra"
)

// submitCmd represents the submit command
func submitCmd(app *app.App) *cobra.Command {
	var (
		lang      string
		timeLimit int
		skipTest  bool
	)
	var (
		problemIDRequired = errors.New("problem id is required")
	)

	cmd := &cobra.Command{
		Use:   "submit",
		Short: "Submit the competitive programming solution",
		Long:  `Run tests, merge libraries, and submit to AtCoder via oj (ongoing contests only).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return problemIDRequired
			}
			return submit.Run(app.Config, submit.Request{
				ProblemPath: args[0],
				Lang:        lang,
				TimeLimit:   timeLimit,
				SkipTest:    skipTest,
			})
		},
	}

	cmd.Flags().StringVarP(&lang, "lang", "l", "cpp", "language of the source code")
	cmd.Flags().IntVarP(&timeLimit, "time-limit", "t", 2, "time limit in seconds for sample tests")
	cmd.Flags().BoolVar(&skipTest, "skip-test", false, "skip sample tests and submit directly")
	return cmd
}
