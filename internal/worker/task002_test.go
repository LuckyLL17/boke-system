package worker

import (
	"context"
	"testing"
	"time"

	"podcast-platform/pkg/logger"
)

type blockingHandler struct {
	started chan struct{}
	release chan struct{}
}

// Worker.Stop -> safeRun -> PublishHandler.Run
func (h *blockingHandler) Name() string { return "blocking-job" }

func (h *blockingHandler) Run(ctx context.Context) error {
	close(h.started)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-h.release:
		return nil
	}
}

func TestShutdownCancelsRunningJob(t *testing.T) {
	log := logger.New(true)
	defer log.Sync()
	w, err := NewWorker(log)
	if err != nil {
		t.Fatal(err)
	}
	h := &blockingHandler{started: make(chan struct{}), release: make(chan struct{})}
	w.AddHandler(h)
	done := make(chan struct{})
	go func() {
		_ = w.RunOnce(h.Name())
		close(done)
	}()
	<-h.started
	_ = w.Stop(context.Background())
	select {
	case <-done:
	case <-time.After(300 * time.Millisecond):
		close(h.release)
		panic("worker shutdown left a running job alive")
	}
}
