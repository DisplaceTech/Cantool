package devserver

import (
	"context"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWatcher_DetectsFileChange(t *testing.T) {
	dir := t.TempDir()
	testFile := filepath.Join(dir, "test.daml")
	require.NoError(t, os.WriteFile(testFile, []byte("initial"), 0644))

	var called atomic.Int32
	w := NewWatcher([]string{dir}, 50*time.Millisecond, func() {
		called.Add(1)
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() { _ = w.Watch(ctx) }()
	time.Sleep(100 * time.Millisecond) // let watcher start

	_ = os.WriteFile(testFile, []byte("changed"), 0644)
	time.Sleep(200 * time.Millisecond) // wait for debounce

	assert.GreaterOrEqual(t, called.Load(), int32(1))
}

func TestWatcher_Debounce(t *testing.T) {
	dir := t.TempDir()
	testFile := filepath.Join(dir, "test.daml")
	require.NoError(t, os.WriteFile(testFile, []byte("initial"), 0644))

	var called atomic.Int32
	w := NewWatcher([]string{dir}, 200*time.Millisecond, func() {
		called.Add(1)
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() { _ = w.Watch(ctx) }()
	time.Sleep(100 * time.Millisecond)

	// Rapid writes should be debounced into one callback
	for i := 0; i < 5; i++ {
		_ = os.WriteFile(testFile, []byte("change"), 0644)
		time.Sleep(10 * time.Millisecond)
	}
	time.Sleep(400 * time.Millisecond)

	assert.Equal(t, int32(1), called.Load())
}

func TestWatcher_ContextCancellation(t *testing.T) {
	dir := t.TempDir()
	w := NewWatcher([]string{dir}, 50*time.Millisecond, func() {})

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- w.Watch(ctx) }()

	cancel()
	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("watcher did not stop on context cancellation")
	}
}

func TestWatcher_NonexistentPath(t *testing.T) {
	w := NewWatcher([]string{"/nonexistent/path"}, 50*time.Millisecond, func() {})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := w.Watch(ctx)
	assert.NoError(t, err)
}

func TestNewWatcher_DefaultDebounce(t *testing.T) {
	w := NewWatcher([]string{}, 0, func() {})
	assert.Equal(t, 100*time.Millisecond, w.debounce)
}
