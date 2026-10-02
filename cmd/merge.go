package cmd

import (
	"errors"
	"fmt"

	"Competitive-Programming-eXecutor/internal/app"
	"Competitive-Programming-eXecutor/internal/merge"
	"Competitive-Programming-eXecutor/internal/problem"

	"github.com/spf13/cobra"
)

func mergeCmd(app *app.App) *cobra.Command {
	var lang string

	cmd := &cobra.Command{
		Use:   "merge <problem>",
		Short: "Merge libraries into a submission file",
		Long:  `Expand library includes in main.<lang> and write the result to submission.<lang>.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("problem id is required")
			}
			p, err := problem.Open(args[0], app.Config.ResolveLang(lang))
			if err != nil {
				return err
			}
			outputPath, err := merge.WriteSubmission(p, app.Config.File.LibraryDirs)
			if err != nil {
				return err
			}
			fmt.Printf("[INFO] wrote %s\n", outputPath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&lang, "lang", "l", "", "language of the source code (default: default_lang in config, else cpp)")

	return cmd
}
