package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/displacetech/cantool/internal/output"
	"github.com/spf13/cobra"
)

var cleanAll bool

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove build artifacts",
	RunE:  runClean,
}

func init() {
	cleanCmd.Flags().BoolVar(&cleanAll, "all", false, "Also remove dist/ and .cantool/ local state")
	rootCmd.AddCommand(cleanCmd)
}

func runClean(cmd *cobra.Command, _ []string) error {
	f := output.New(Format())

	targets := []string{".daml"}
	if cleanAll {
		targets = append(targets, "dist", ".cantool")
	}

	removed := 0
	for _, dir := range targets {
		abs, _ := filepath.Abs(dir)
		info, err := os.Stat(abs)
		if err != nil {
			continue
		}
		if !info.IsDir() {
			continue
		}
		size := dirSize(abs)
		if err := os.RemoveAll(abs); err != nil {
			f.Error(output.Errorf("CT4002", "Check file permissions",
				"failed to remove %s", dir))
			continue
		}
		f.Info(fmt.Sprintf("  Removed: %s (%s)", dir, humanSize(size)))
		removed++
	}

	if removed == 0 {
		f.Info("Nothing to clean — no build artifacts found.")
	} else if cleanAll {
		f.Success("Cleaned all artifacts", nil)
	} else {
		f.Success("Cleaned build artifacts", nil)
	}

	return nil
}

func dirSize(path string) int64 {
	var size int64
	_ = filepath.Walk(path, func(_ string, info os.FileInfo, _ error) error {
		if info != nil && !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
}

func humanSize(b int64) string {
	const (
		kb = 1024
		mb = 1024 * kb
	)
	switch {
	case b >= mb:
		return fmt.Sprintf("%d MB", b/mb)
	case b >= kb:
		return fmt.Sprintf("%d KB", b/kb)
	default:
		return fmt.Sprintf("%d B", b)
	}
}
