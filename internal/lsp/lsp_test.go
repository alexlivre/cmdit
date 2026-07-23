package lsp

import (
	"testing"
	"time"
)

func TestClientShutdownStopsReadLoop(t *testing.T) {
	c, err := NewClient("sleep", "60")
	if err != nil {
		t.Skipf("could not start test process: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	done := make(chan struct{})
	go func() {
		c.Shutdown()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Shutdown timed out — readLoop may not be stopping")
	}
}

func TestContextCancelExitsReadLoop(t *testing.T) {
	c, err := NewClient("sleep", "60")
	if err != nil {
		t.Skipf("could not start test process: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	c.cancel()

	select {
	case <-c.done:
	case <-time.After(3 * time.Second):
		t.Fatal("context cancel did not stop readLoop")
	}
}
