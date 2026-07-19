/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"Competitive-Programming-eXecutor/internal/app"

	"github.com/spf13/cobra"
)

// newLibCmd represents the new-lib command
func newLibCmd(app *app.App) *cobra.Command {
	var dirIndex int
	cmd := &cobra.Command{
		Use:   "new-lib <path>",
		Short: "Create a library file from the template",
		Long: `Create a new library header under config.library_dirs.

<path> is relative to the selected library directory (e.g. graph/dsu.hpp).
The file is copied from .cpx/templates/library/library_template.hpp.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			libraryDirs := app.Config.File.LibraryDirs
			if len(libraryDirs) == 0 {
				return fmt.Errorf("library_dirs is empty: set it in .cpx/config.yaml")
			}
			if dirIndex < 0 || dirIndex >= len(libraryDirs) {
				return fmt.Errorf("invalid --dir index %d: library_dirs has %d entries (0-%d)",
					dirIndex, len(libraryDirs), len(libraryDirs)-1)
			}

			relPath := filepath.Clean(args[0])
			if relPath == "." || strings.HasPrefix(relPath, "..") || filepath.IsAbs(relPath) {
				return fmt.Errorf("path must be relative to the library directory: %q", args[0])
			}

			libraryDir := libraryDirs[dirIndex]
			filePath := filepath.Join(libraryDir, relPath)
			if _, err := os.Stat(filePath); err == nil {
				return fmt.Errorf("library already exists: %s", filePath)
			} else if err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("stat %q: %w", filePath, err)
			}

			if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
				return fmt.Errorf("create library directory: %w", err)
			}

			templatePath := filepath.Join(
				app.Config.File.RootDir,
				".cpx", "templates", "library", "library_template.hpp",
			)
			src, err := os.Open(templatePath)
			if err != nil {
				if os.IsNotExist(err) {
					return fmt.Errorf("library template not found: %s (run `cpx init`)", templatePath)
				}
				return fmt.Errorf("open template %q: %w", templatePath, err)
			}
			defer src.Close()

			dst, err := os.Create(filePath)
			if err != nil {
				return fmt.Errorf("create library %q: %w", filePath, err)
			}
			defer dst.Close()

			if _, err := io.Copy(dst, src); err != nil {
				return fmt.Errorf("copy template to %q: %w", filePath, err)
			}

			includePath := filepath.ToSlash(relPath)
			fmt.Printf("[INFO] created %s\n", filePath)
			fmt.Printf("[INFO] include: #include %q\n", includePath)
			return nil
		},
	}
	cmd.Flags().IntVarP(&dirIndex, "dir", "d", 0, "index into config.library_dirs (default 0)")
	return cmd
}
