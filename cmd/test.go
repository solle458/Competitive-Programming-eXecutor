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
		timeLimit float64
		eps       float64
	)

	cmd := &cobra.Command{
		Use:   "test <problem>",
		Short: "Run sample tests for a problem",
		Long:  `Compile the solution and compare its output against sample cases under the problem directory.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("problem id is required")
			}
			if timeLimit <= 0 {
				return errors.New("time limit must be a positive number of seconds")
			}
			p, err := problem.Open(args[0], app.Config.ResolveLang(lang))
			if err != nil {
				return err
			}
			err = test.RunSamples(p, test.Options{TimeLimit: timeLimit, Eps: eps}, app.Config)
			if errors.Is(err, test.ErrNotAccepted) {
				cmd.SilenceUsage = true
			}
			return err
		},
	}

	cmd.Flags().StringVarP(&lang, "lang", "l", "", "language of the source code (default: default_lang in config, else cpp)")
	cmd.Flags().Float64VarP(&timeLimit, "time-limit", "t", 2, "time limit in seconds for sample tests (fractions such as 1.5 are allowed)")
	cmd.Flags().Float64Var(&eps, "eps", 0, "absolute or relative error allowed for floating-point outputs")
	return cmd
}
