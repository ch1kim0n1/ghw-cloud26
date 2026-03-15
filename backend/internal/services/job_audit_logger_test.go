package services

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"
)

type flakyAuditLogger struct {
	mu          sync.Mutex
	failUntil   int
	calls       int
	healthValue AuditHealth
}

func (f *flakyAuditLogger) Record(context.Context, JobAuditEvent) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	if f.calls <= f.failUntil {
		return errors.New("temporary failure")
	}
	return nil
}

func (f *flakyAuditLogger) Health(context.Context) AuditHealth {
	if f.healthValue.Status == "" {
		return AuditHealth{Enabled: true, Status: "healthy", Details: "ok"}
	}
	return f.healthValue
}

func (f *flakyAuditLogger) Calls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls
}

func TestAsyncJobAuditLoggerRetriesTransientErrors(t *testing.T) {
	base := &flakyAuditLogger{failUntil: 2}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	audit := NewAsyncJobAuditLogger(base, logger)
	defer audit.Close()

	if err := audit.Record(context.Background(), JobAuditEvent{JobID: "job_1", EventType: "job_transitioned"}); err != nil {
		t.Fatalf("Record() error = %v", err)
	}
	audit.Wait()

	if got := base.Calls(); got != 3 {
		t.Fatalf("expected 3 record attempts, got %d", got)
	}
}

func TestAsyncJobAuditLoggerHealthDelegatesToBase(t *testing.T) {
	expected := AuditHealth{Enabled: true, Status: "healthy", Details: "notion audit sink connected"}
	base := &flakyAuditLogger{healthValue: expected}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	audit := NewAsyncJobAuditLogger(base, logger)
	defer audit.Close()

	health := audit.Health(context.Background())
	if health != expected {
		t.Fatalf("unexpected health payload: got %#v want %#v", health, expected)
	}
}

func TestNoopAuditLoggerHealthIsDisabled(t *testing.T) {
	logger := NewNoopJobAuditLogger()
	health := logger.Health(context.Background())
	if health.Enabled {
		t.Fatal("expected noop logger health to be disabled")
	}
	if health.Status != "disabled" {
		t.Fatalf("expected disabled status, got %q", health.Status)
	}
}

func TestAsyncJobAuditLoggerHandlesContextCancellation(t *testing.T) {
	base := &flakyAuditLogger{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	audit := NewAsyncJobAuditLogger(base, logger)
	defer audit.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := audit.Record(ctx, JobAuditEvent{JobID: "job_ctx", EventType: "cancel_test", Timestamp: time.Now()}); err != nil {
		t.Fatalf("Record() error = %v", err)
	}
	audit.Wait()
	if got := base.Calls(); got != 1 {
		t.Fatalf("expected one write attempt, got %d", got)
	}
}
