/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"Competitive-Programming-eXecutor/internal/app"
	"Competitive-Programming-eXecutor/internal/setup"
	"Competitive-Programming-eXecutor/internal/setup/atcoder"
	"errors"
	"os"

	"github.com/spf13/cobra"
)

// setupCmd represents the setup command
func setupCmd(app *app.App) *cobra.Command {
	var lang string
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Setup the competitive programming",
		Long:  `Setup the competitive programming`,
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
			})
			if err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&lang, "lang", "l", "cpp", "language of the source code")
	return cmd
}
