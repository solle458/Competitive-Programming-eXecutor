/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"Competitive-Programming-eXecutor/internal/app"
	"Competitive-Programming-eXecutor/internal/config"
	"Competitive-Programming-eXecutor/internal/template"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// initCmd represents the init command
func initCmd(app *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize the competitive programming",
		Long:  `Initialize the competitive programming`,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := os.Getwd()
			if err != nil {
				return err
			}
			app.Config.File.RootDir = root
			app.Config.File.LibraryDirs = []string{filepath.Join(root, "library")}
			app.Config.File.DefaultLang = "cpp"
			app.Config.File.AtCoderSession = ""
			if err := config.Update(app.Config); err != nil {
				return err
			}
			if err := template.CreateTemplate(app.Config.File.RootDir); err != nil {
				return err
			}
			if err := template.CreateLibraryTemplate(app.Config.File.RootDir); err != nil {
				return err
			}
			return nil
		},
	}
}
