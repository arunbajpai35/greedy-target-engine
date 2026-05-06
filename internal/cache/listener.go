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
	reload := make(chan struct{}, 1)
	trigger := func(reason string) {
		slog.Info("cache reload requested", "reason", reason)
		select {
		case reload <- struct{}{}:
		default:
		}
	}

	l := pq.NewListener(connStr, 5*time.Second, 30*time.Second, func(ev pq.ListenerEventType, err error) {
		switch ev {
		case pq.ListenerEventConnected, pq.ListenerEventReconnected:
			trigger("listener (re)connected")
		case pq.ListenerEventDisconnected:
			slog.Warn("listener disconnected", "err", err)
		case pq.ListenerEventConnectionAttemptFailed:
			slog.Warn("listener connect attempt failed", "err", err)
		}
	})
	if err := l.Listen(NotifyChannel); err != nil {
		return err
	}

	go func() {
		defer l.Close()
		var pending bool
		var timer *time.Timer
		ping := time.NewTicker(20 * time.Second)
		defer ping.Stop()

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

		schedule := func() {
			if pending {
				return
			}
			pending = true
			timer = time.AfterFunc(debounce, fire)
		}

		for {
			select {
			case <-ctx.Done():
				if timer != nil {
					timer.Stop()
				}
				return
			case <-l.Notify:
				schedule()
			case <-reload:
				schedule()
			case <-ping.C:
				_ = l.Ping()
			}
		}
	}()
	return nil
}
