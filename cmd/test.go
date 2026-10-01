package cmd

import (
	"errors"

	"Competitive-Programming-eXecutor/internal/app"
	"Competitive-Programming-eXecutor/internal/problem"
	"Competitive-Programming-eXecutor/internal/test"

	"github.com/spf13/cobra"
)

func testCmd(app *app.App) *cobra.Command {
	var (
		lang      string
		timeLimit int
	)

	cmd := &cobra.Command{
		Use:   "test <problem>",
		Short: "Run sample tests for a problem",
		Long:  `Compile the solution and compare its output against sample cases under the problem directory.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("problem id is required")
			}
			return test.RunSamples(problem.Problem{Dir: args[0], Lang: lang}, timeLimit, app.Config)
		},
	}

	cmd.Flags().StringVarP(&lang, "lang", "l", app.Config.File.DefaultLang, "language of the source code")
	cmd.Flags().IntVarP(&timeLimit, "time-limit", "t", 2, "time limit in seconds for sample tests")
	return cmd
}
