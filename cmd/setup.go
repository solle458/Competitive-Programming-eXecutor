/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"Competitive-Programming-eXecutor/internal/app"
	"Competitive-Programming-eXecutor/internal/setup"
	"Competitive-Programming-eXecutor/internal/setup/atcoder"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// setupCmd represents the setup command
func setupCmd(app *app.App) *cobra.Command {
	var lang string
	cmd := &cobra.Command{
		Use:   "setup <contest>",
		Short: "Download contest problems and sample cases",
		Long:  `Create problem directories for an AtCoder contest, with templates and sample input/output files.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("contest id is required")
			}
			contestID := args[0]
			atcoder := atcoder.AtCoder{}
			exists := atcoder.Supports(contestID)
			if !exists {
				return errors.New("contest not found")
			}
			wd, err := os.Getwd()
			if err != nil {
				return err
			}
			err = atcoder.Setup(setup.Request{
				ContestID:  contestID,
				Lang:       lang,
				WorkingDir: wd,
				Config:     app.Config,
			}, app)
			if err != nil {
				return err
			}
			fmt.Printf("[INFO] setup complete for contest %s\n", contestID)
			return nil
		},
	}
	cmd.Flags().StringVarP(&lang, "lang", "l", app.Config.File.DefaultLang, "language of the source code")
	return cmd
}
