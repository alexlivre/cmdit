package lsp

import (
	"os"
	"os/exec"
	"runtime"
	"testing"
	"time"
)

// hangCommand returns a long-running process with quiet stdout so readLoop
// blocks on I/O until pipes are closed / the process is killed.
func hangCommand() (name string, args []string) {
	if runtime.GOOS == "windows" {
		return "powershell", []string{"-NoProfile", "-NonInteractive", "-Command", "Start-Sleep -Seconds 60"}
	}
	return "sleep", []string{"60"}
}

func startHangClient(t *testing.T) *Client {
	t.Helper()
	name, args := hangCommand()
	if _, err := exec.LookPath(name); err != nil {
		t.Skipf("hang helper %q not available: %v", name, err)
	}
	c, err := NewClient(name, args...)
	if err != nil {
		t.Skipf("could not start test process: %v", err)
	}
	t.Cleanup(func() {
		c.forceClose()
		select {
		case <-c.done:
		case <-time.After(2 * time.Second):
		}
	})
	return c
}

func TestClientShutdownStopsReadLoop(t *testing.T) {
	c := startHangClient(t)
	time.Sleep(100 * time.Millisecond)

	done := make(chan struct{})
	go func() {
		_ = c.Shutdown()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Shutdown timed out — readLoop may not be stopping")
	}
}

func TestContextCancelExitsReadLoop(t *testing.T) {
	c := startHangClient(t)
	time.Sleep(100 * time.Millisecond)

	c.cancel()

	select {
	case <-c.done:
	case <-time.After(3 * time.Second):
		t.Fatal("context cancel did not stop readLoop")
	}
}

func TestShutdownIdempotent(t *testing.T) {
	c := startHangClient(t)
	time.Sleep(50 * time.Millisecond)
	if err := c.Shutdown(); err != nil {
		t.Fatalf("first Shutdown: %v", err)
	}
	// Second call must not hang or panic.
	done := make(chan struct{})
	go func() {
		_ = c.Shutdown()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("second Shutdown hung")
	}
}

// Ensure the hang helper itself exists in CI environments.
func TestHangHelperAvailable(t *testing.T) {
	name, _ := hangCommand()
	path, err := exec.LookPath(name)
	if err != nil {
		t.Fatalf("expected hang helper %q on %s: %v (PATH=%s)", name, runtime.GOOS, err, os.Getenv("PATH"))
	}
	t.Logf("hang helper: %s", path)
}
