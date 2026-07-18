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
		copy      bool
	)
	var (
		problemIDRequired = errors.New("problem id is required")
	)

	cmd := &cobra.Command{
		Use:   "submit <problem>",
		Short: "Test, merge, and submit a solution",
		Long: `Run sample tests, merge libraries into a submission file, then submit to AtCoder via oj
(with a MiB memory-limit patch so modern AtCoder pages parse correctly).

Use --copy to copy the merged source to the clipboard instead of submitting
(useful for practice on past problems). Live oj submit works for ongoing contests only.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return problemIDRequired
			}
			return submit.Run(app.Config, submit.Request{
				ProblemPath: args[0],
				Lang:        lang,
				TimeLimit:   timeLimit,
				SkipTest:    skipTest,
				Copy:        copy,
			})
		},
	}

	cmd.Flags().StringVarP(&lang, "lang", "l", app.Config.File.DefaultLang, "language of the source code")
	cmd.Flags().IntVarP(&timeLimit, "time-limit", "t", 2, "time limit in seconds for sample tests")
	cmd.Flags().BoolVar(&skipTest, "skip-test", false, "skip sample tests before submit or copy")
	cmd.Flags().BoolVarP(&copy, "copy", "c", false, "copy merged source to clipboard instead of submitting")
	return cmd
}
