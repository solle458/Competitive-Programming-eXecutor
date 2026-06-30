/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"Competitive-Programming-eXecutor/internal/app"
	"Competitive-Programming-eXecutor/internal/merge"

	"github.com/spf13/cobra"
)

// mergeCmd represents the merge command
func mergeCmd(app *app.App) *cobra.Command {
	var lang string

	cmd := &cobra.Command{
		Use:   "merge",
		Short: "Merge the competitive programming",
		Long:  `Merge the competitive programming`,
		RunE: func(cmd *cobra.Command, args []string) error {
			extension := lang
			if extension == "" {
				extension = app.Config.File.DefaultLang
			}
			if len(args) < 1 {
				return errors.New("problem id is required")
			}
			problemID := args[0]
			if err := validProblemID(problemID); err != nil {
				return err
			}

			sourceCode, err := getSourceCode(problemID, extension)
			if err != nil {
				return err
			}
			submissionCode, err := merge.Generate(sourceCode, app.Config.File.LibraryDirs)
			if err != nil {
				return err
			}

			outputPath := filepath.Join(".", problemID, "submission."+extension)
			if err := os.WriteFile(outputPath, []byte(submissionCode), 0o644); err != nil {
				return fmt.Errorf("write submission file %q: %w", outputPath, err)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&lang, "lang", "l", "", "language of the source code")

	return cmd
}

func getSourceCode(id, extension string) (string, error) {
	sourceCodePath := filepath.Join(".", id, "main."+extension)
	content, err := os.ReadFile(sourceCodePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("source code not found: %q", sourceCodePath)
		}
		return "", fmt.Errorf("read source code %q: %w", sourceCodePath, err)
	}
	return string(content), nil
}

func validProblemID(id string) error {
	problemDir := filepath.Join(".", id)
	info, err := os.Stat(problemDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("problem %q not found in current directory", id)
		}
		return fmt.Errorf("stat problem directory %q: %w", problemDir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%q is not a directory", problemDir)
	}
	return nil
}
