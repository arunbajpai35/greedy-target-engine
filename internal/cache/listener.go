package cache

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/lib/pq"
)

const NotifyChannel = "targeting_changes"

func StartListener(ctx context.Context, c *Cache, db *sql.DB, connStr string, debounce time.Duration) error {
	l := pq.NewListener(connStr, 5*time.Second, time.Minute, func(ev pq.ListenerEventType, err error) {
		if err != nil {
			slog.Warn("listener event", "type", ev, "err", err)
		}
	})
	if err := l.Listen(NotifyChannel); err != nil {
		return err
	}

	go func() {
		defer l.Close()
		var pending bool
		var timer *time.Timer

		fire := func() {
			pending = false
			rctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			if err := c.Reload(rctx, db); err != nil {
				slog.Error("cache reload", "err", err)
			} else {
				slog.Info("cache reloaded", "size", c.Size())
			}
			cancel()
		}

		for {
			select {
			case <-ctx.Done():
				if timer != nil {
					timer.Stop()
				}
				return
			case <-l.Notify:
				if pending {
					continue
				}
				pending = true
				timer = time.AfterFunc(debounce, fire)
			case <-time.After(time.Minute):
				_ = l.Ping()
			}
		}
	}()
	return nil
}
