// Package jobs holds the queued jobs.
package jobs

import (
	"context"
	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/queue"
	"time"
)

// All lists the job handlers, built from the app so they reach its
// database, clock and mailer; vukem make:job appends here and names the
// parameter "app" with the first handler that needs it.
func All(app *vukem.App) []queue.Registration {
	return []queue.Registration{queue.Handle(func(ctx context.Context, j SendDelivery) error {
		return deliver(app, ctx, j)
	}, queue.Retries(5), queue.Backoff(queue.Exponential(30*time.Second)), queue.Timeout(45*time.Second))}
}
