package devserver

import (
	"context"
	"fmt"
	"time"

	"github.com/displacetech/cantool/internal/config"
	"github.com/displacetech/cantool/internal/exec"
	"github.com/displacetech/cantool/internal/ledger"
	"github.com/displacetech/cantool/internal/output"
	"github.com/displacetech/cantool/internal/sdk"
)

// DevServer orchestrates the local development loop.
type DevServer struct {
	Config    *config.Config
	Runner    exec.CommandRunner
	Formatter output.Formatter
	Port      int
}

// Start runs the full dev lifecycle: build → sandbox → health → parties → watch.
// It blocks until ctx is cancelled.
func (d *DevServer) Start(ctx context.Context) error {
	// 1. Detect SDK
	detector := &sdk.PathDetector{Runner: d.Runner}
	sdkInfo, err := detector.Detect(ctx)
	if err != nil {
		if ce, ok := err.(*output.CantoolError); ok {
			d.Formatter.Error(ce)
		}
		return err
	}

	// 2. Build
	d.Formatter.Info("Building project...")
	result, err := sdk.Build(ctx, d.Runner, sdkInfo, ".")
	if err != nil {
		if ce, ok := err.(*output.CantoolError); ok {
			d.Formatter.Error(ce)
		}
		return err
	}
	if !result.Success {
		d.Formatter.Error(output.Errorf("CT4001", "Fix build errors", "build failed"))
		return fmt.Errorf("build failed")
	}
	d.Formatter.Success(fmt.Sprintf("Build succeeded (%s)", result.Duration.Round(time.Millisecond)), nil)

	// 3. Start sandbox
	port := d.Port
	if port == 0 {
		port = d.Config.Dev.SandboxPort
	}
	if port == 0 {
		port = 5011
	}

	d.Formatter.Info(fmt.Sprintf("Starting sandbox on port %d...", port))
	sandbox, err := sdk.StartSandbox(ctx, sdkInfo, sdk.SandboxConfig{
		Port:       port,
		DARPath:    result.DARPath,
		ProjectDir: ".",
	})
	if err != nil {
		if ce, ok := err.(*output.CantoolError); ok {
			d.Formatter.Error(ce)
		}
		return err
	}
	defer sandbox.Stop()

	// 4. Health poll
	baseURL := fmt.Sprintf("http://localhost:%d", port)
	client := ledger.NewHTTPLedgerClient(baseURL)
	if err := d.waitForHealth(ctx, client); err != nil {
		return err
	}
	d.Formatter.Success("Sandbox ready", nil)

	// 5. Allocate parties
	d.Formatter.Info("Allocating parties...")
	for _, p := range d.Config.Parties {
		_, err := client.AllocateParty(ctx, p.Name, p.Display)
		if err != nil {
			d.Formatter.Warning(fmt.Sprintf("Party %s: %v (non-fatal)", p.Name, err))
		} else {
			d.Formatter.Info(fmt.Sprintf("  ✓ %s (%s)", p.Name, p.Display))
		}
	}

	// 6. Watch for changes
	watchPaths := d.Config.Dev.WatchPaths
	if len(watchPaths) == 0 {
		watchPaths = []string{"daml/"}
	}
	d.Formatter.Info(fmt.Sprintf("\nWatching %v for changes. Press Ctrl+C to stop.", watchPaths))

	watcher := NewWatcher(watchPaths, 100*time.Millisecond, func() {
		d.Formatter.Info(fmt.Sprintf("\n[%s] Rebuilding...", time.Now().Format("15:04:05")))
		r, err := sdk.Build(ctx, d.Runner, sdkInfo, ".")
		if err != nil || !r.Success {
			d.Formatter.Error(output.Errorf("CT4001", "Fix build errors", "rebuild failed"))
			return
		}
		d.Formatter.Success("Build succeeded, uploading DAR...", nil)
		if r.DARPath != "" {
			if err := client.UploadDAR(ctx, r.DARPath); err != nil {
				d.Formatter.Warning(fmt.Sprintf("DAR upload: %v", err))
			} else {
				d.Formatter.Success("DAR uploaded", nil)
			}
		}
	})

	return watcher.Watch(ctx)
}

func (d *DevServer) waitForHealth(ctx context.Context, client ledger.LedgerClient) error {
	d.Formatter.Info("Waiting for sandbox to be ready...")
	deadline := time.After(60 * time.Second)
	backoff := 500 * time.Millisecond

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-deadline:
			return output.Errorf("CT3001",
				"The sandbox did not become ready within 60 seconds. Check Java and SDK installation",
				"sandbox health check timed out")
		default:
			h, err := client.Health(ctx)
			if err == nil && h.Status == "SERVING" {
				return nil
			}
			time.Sleep(backoff)
			if backoff < 5*time.Second {
				backoff = backoff * 2
			}
		}
	}
}
