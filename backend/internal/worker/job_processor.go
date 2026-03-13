package worker

import (
	"context"
	"log/slog"
	"sync/atomic"
	"time"
)

type Processor struct {
	logger   *slog.Logger
	interval time.Duration
	ticks    atomic.Int64
	onTick   func()
}

func NewProcessor(logger *slog.Logger, interval time.Duration) *Processor {
	return &Processor{
		logger:   logger,
		interval: interval,
	}
}

func (p *Processor) Run(ctx context.Context) {
	p.logger.Info("worker stub started", "interval", p.interval.String())
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			p.logger.Info("worker stub stopped", "ticks", p.TickCount())
			return
		case <-ticker.C:
			p.ticks.Add(1)
			p.logger.Info("worker stub tick")
			if p.onTick != nil {
				p.onTick()
			}
		}
	}
}

func (p *Processor) SetOnTick(fn func()) {
	p.onTick = fn
}

func (p *Processor) TickCount() int64 {
	return p.ticks.Load()
}
