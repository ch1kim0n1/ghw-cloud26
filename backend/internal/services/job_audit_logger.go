package services

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/models"
)

type JobAuditEvent struct {
	JobID        string
	CampaignID   string
	Status       string
	CurrentStage string
	EventType    string
	Message      string
	ErrorCode    string
	Timestamp    time.Time
	Metadata     models.Metadata
}

type AuditHealth struct {
	Enabled bool
	Status  string
	Details string
}

type JobAuditLogger interface {
	Record(context.Context, JobAuditEvent) error
	Health(context.Context) AuditHealth
}

type NoopJobAuditLogger struct{}

func NewNoopJobAuditLogger() *NoopJobAuditLogger {
	return &NoopJobAuditLogger{}
}

func (n *NoopJobAuditLogger) Record(context.Context, JobAuditEvent) error {
	return nil
}

func (n *NoopJobAuditLogger) Health(context.Context) AuditHealth {
	return AuditHealth{Enabled: false, Status: "disabled", Details: "audit sink is not configured"}
}

type queuedAuditEvent struct {
	event JobAuditEvent
	ctx   context.Context
}

type AsyncJobAuditLogger struct {
	base        JobAuditLogger
	logger      *slog.Logger
	queue       chan queuedAuditEvent
	maxAttempts int
	wg          sync.WaitGroup
	closeOnce   sync.Once
}

func NewAsyncJobAuditLogger(base JobAuditLogger, logger *slog.Logger) *AsyncJobAuditLogger {
	if base == nil {
		base = NewNoopJobAuditLogger()
	}
	a := &AsyncJobAuditLogger{
		base:        base,
		logger:      logger,
		queue:       make(chan queuedAuditEvent, 128),
		maxAttempts: 3,
	}
	go a.run()
	return a
}

func (a *AsyncJobAuditLogger) Record(ctx context.Context, event JobAuditEvent) error {
	if ctx == nil {
		ctx = context.Background()
	}
	a.wg.Add(1)
	item := queuedAuditEvent{event: event, ctx: ctx}
	select {
	case a.queue <- item:
	default:
		go a.process(item)
	}
	return nil
}

func (a *AsyncJobAuditLogger) Health(ctx context.Context) AuditHealth {
	return a.base.Health(ctx)
}

func (a *AsyncJobAuditLogger) Close() {
	a.closeOnce.Do(func() {
		close(a.queue)
	})
}

func (a *AsyncJobAuditLogger) Wait() {
	a.wg.Wait()
}

func (a *AsyncJobAuditLogger) run() {
	for item := range a.queue {
		a.process(item)
	}
}

func (a *AsyncJobAuditLogger) process(item queuedAuditEvent) {
	defer a.wg.Done()

	var lastErr error
	for attempt := 1; attempt <= a.maxAttempts; attempt++ {
		err := a.base.Record(item.ctx, item.event)
		if err == nil {
			return
		}
		lastErr = err
		if attempt < a.maxAttempts {
			time.Sleep(time.Duration(attempt*200) * time.Millisecond)
		}
	}

	if a.logger != nil {
		a.logger.Warn("job audit logger write failed after retries", "job_id", item.event.JobID, "event_type", item.event.EventType, "attempts", a.maxAttempts, "error", lastErr)
	}
}
