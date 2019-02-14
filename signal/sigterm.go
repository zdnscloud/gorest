package signal

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func SigTermCancelContext(ctx context.Context) context.Context {
	term := make(chan os.Signal)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(ctx)

	go func() {
		select {
		case <-term:
			cancel()
		case <-ctx.Done():
		}
	}()

	return ctx
}
