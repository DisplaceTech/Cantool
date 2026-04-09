package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/displacetech/cantool/internal/config"
	"github.com/displacetech/cantool/internal/convenience"
	"github.com/displacetech/cantool/internal/exec"
	"github.com/displacetech/cantool/internal/output"
	"github.com/displacetech/cantool/internal/sdk"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

var BuildWatch bool

var BuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Compile DAML sources",
	RunE:  runBuild,
}

func init() {
	BuildCmd.Flags().BoolVar(&BuildWatch, "watch", false, "Watch for changes and rebuild")
}

func runBuild(cmd *cobra.Command, _ []string) error {
	f := output.New(Format())
	runner := &exec.DefaultRunner{}
	ctx := cmd.Context()

	cfg, err := config.Load()
	if err != nil {
		f.Error(err.(*output.CantoolError))
		return err
	}

	detector := &sdk.PathDetector{Runner: runner}
	sdkInfo, err := detector.Detect(ctx)
	if err != nil {
		if ce, ok := err.(*output.CantoolError); ok {
			f.Error(ce)
		}
		return err
	}

	convenience.PrintDelegation(sdkInfo.Command, "build")

	if err := doBuild(ctx, f, runner, sdkInfo, cfg); err != nil {
		return err
	}

	if BuildWatch {
		return watchAndRebuild(ctx, f, runner, sdkInfo, cfg)
	}

	return nil
}

func doBuild(ctx context.Context, f output.Formatter, runner exec.CommandRunner, sdkInfo *sdk.SDKInfo, cfg *config.Config) error {
	result, err := sdk.Build(ctx, runner, sdkInfo, ".")
	if err != nil {
		if ce, ok := err.(*output.CantoolError); ok {
			f.Error(ce)
		}
		return err
	}

	if !result.Success {
		f.Error(output.Errorf("CT4001", "Fix the errors above",
			"Build failed (%s)", result.Duration.Round(time.Millisecond)))
		for _, e := range result.Errors {
			f.Info("  " + e)
		}
		return fmt.Errorf("build failed")
	}

	msg := fmt.Sprintf("Build succeeded (%s)", result.Duration.Round(time.Millisecond))
	f.Success(msg, map[string]any{
		"success":     true,
		"dar_path":    result.DARPath,
		"duration_ms": result.Duration.Milliseconds(),
	})
	if result.DARPath != "" {
		f.Info(fmt.Sprintf("  DAR: %s", result.DARPath))
	}
	for _, w := range result.Warnings {
		f.Warning(w)
	}
	_ = cfg // used for watch paths
	return nil
}

func watchAndRebuild(ctx context.Context, f output.Formatter, runner exec.CommandRunner, sdkInfo *sdk.SDKInfo, cfg *config.Config) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer func() { _ = watcher.Close() }()

	watchPaths := cfg.Dev.WatchPaths
	if len(watchPaths) == 0 {
		watchPaths = []string{"daml/"}
	}
	for _, p := range watchPaths {
		_ = watcher.Add(p)
	}

	f.Info("Watching for changes. Press Ctrl+C to stop.")

	var debounce *time.Timer
	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if event.Op&(fsnotify.Write|fsnotify.Create) == 0 {
				continue
			}
			if debounce != nil {
				debounce.Stop()
			}
			debounce = time.AfterFunc(100*time.Millisecond, func() {
				f.Info(fmt.Sprintf("\n[%s] File changed: %s", time.Now().Format("15:04:05"), event.Name))
				f.Info("Rebuilding...")
				_ = doBuild(ctx, f, runner, sdkInfo, cfg)
			})
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			f.Warning(fmt.Sprintf("Watch error: %v", err))
		}
	}
}
