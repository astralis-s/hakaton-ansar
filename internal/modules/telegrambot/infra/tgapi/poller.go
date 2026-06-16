package tgapi

import (
	"context"
	"log/slog"
	"time"
)

// Poller long-polls getUpdates and dispatches each update to handle, until ctx
// is cancelled. Network errors are logged and retried after a short backoff.
type Poller struct {
	client     *Client
	log        *slog.Logger
	timeoutSec int
}

func NewPoller(client *Client, log *slog.Logger, timeoutSec int) *Poller {
	if timeoutSec <= 0 {
		timeoutSec = 30
	}
	return &Poller{client: client, log: log, timeoutSec: timeoutSec}
}

// Run blocks, processing updates until ctx is cancelled.
func (p *Poller) Run(ctx context.Context, handle func(context.Context, Update)) {
	var offset int64
	for {
		if ctx.Err() != nil {
			return
		}
		updates, err := p.client.GetUpdates(ctx, offset, p.timeoutSec)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			p.log.Error("telegram getUpdates failed", "error", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(3 * time.Second):
			}
			continue
		}
		for _, u := range updates {
			if u.UpdateID >= offset {
				offset = u.UpdateID + 1
			}
			handle(ctx, u)
		}
	}
}
