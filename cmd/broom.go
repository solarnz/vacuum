package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/karrick/tparse"
	"github.com/spf13/cobra"
)

func isEmptyDir(file *os.File) (bool, error) {
	_, err := file.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return false, nil
}

func vacuum(cmd *cobra.Command, args []string) {
	dryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to parse dry-run: %s", err)
		os.Exit(1)
	}

	durationString, err := cmd.Flags().GetString("duration")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to get duration: %s", err)
		os.Exit(1)
	}

	duration, err := tparse.ParseNow(time.RFC3339, fmt.Sprintf("now-%s", durationString))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	_, _ = dryRun, duration

	for _, rootPath := range args {
		err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if rootPath == path {
				return nil
			}

			if !info.IsDir() && info.ModTime().Before(duration) {
				if dryRun {
					fmt.Printf("Would remove %s\n", path)
				} else {
					err = os.Remove(path)
					if err != nil {
						return err
					}
					fmt.Printf("Removed %s\n", path)
					return nil
				}
			}

			if info.IsDir() {
				file, err := os.Open(path)
				if err != nil {
					return err
				}
				empty, err := isEmptyDir(file)
				if err != nil {
					return err
				}
				if empty {
					if dryRun {
						fmt.Printf("Would remove %s\n", path)
					} else {
						fmt.Printf("Removing empty directory %s\n", path)
						os.Remove(path)
						return nil
					}
				}
			}
			return nil
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error walking directory %s: %s\n", rootPath, err)
			os.Exit(1)
		}
	}
}
