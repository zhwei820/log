package log

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/zhwei820/sentry-grpc/sentryclient"
	"go.uber.org/zap"
)

func SentryClient(sentryAddr string) *sentry.Client {
	client, _ := sentry.NewClient(sentry.ClientOptions{
		Dsn: "",
		Transport: &transport{
			sentryclient: sentryclient.NewSentryCli(context.Background(), sentryAddr),
		},
	})
	return client
}

type transport struct {
	sentryclient sentryclient.SentryClient
}

// Flush waits until any buffered events are sent to the Sentry server, blocking
// for at most the given timeout. It returns false if the timeout was reached.
func (f *transport) Flush(_ time.Duration) bool { return true }

// Configure is called by the Client itself, providing it it's own ClientOptions.
func (f *transport) Configure(_ sentry.ClientOptions) {}

// SendEvent assembles a new packet out of Event and sends it to remote server.
// We use this method to capture the event for testing
func (f *transport) SendEvent(event *sentry.Event) {
	ctx := context.Background()
	DebugZ(ctx, "sentryclient.SendEvent start")

	go func() {
		if _, err := f.sentryclient.SendEvent(ctx, sentryclient.FromSentryEvent(event)); err != nil {
			WarnZ(ctx, "sentryclient.SendEvent failed", zap.Error(err))
		}
	}()
}
