package telegram

import (
	"context"
	"log/slog"
	"time"
)

// Poller крутит long polling getUpdates и передаёт каждое обновление в handle.
// Корректно завершается по отмене контекста.
type Poller struct {
	client     *Client
	log        *slog.Logger
	timeoutSec int
}

// NewPoller создаёт поллер с указанным таймаутом long polling (в секундах).
func NewPoller(client *Client, log *slog.Logger, timeoutSec int) *Poller {
	if timeoutSec <= 0 {
		timeoutSec = 30
	}
	return &Poller{client: client, log: log, timeoutSec: timeoutSec}
}

// Run блокирующе обрабатывает обновления до отмены ctx. Ошибки сети логируются,
// после чего следует короткий бэкофф и повтор.
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
			p.log.Error("getUpdates failed", "error", err)
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
