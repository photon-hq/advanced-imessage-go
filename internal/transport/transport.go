// Package transport builds the Connect/gRPC plumbing for the imessage client:
// the HTTP/2 client, base URL, and the interceptor chain (auth, timeout,
// idempotency, retry). It is deliberately free of any domain types so it can be
// imported by the root package without a cycle; error translation to the
// public error type happens at the root-package boundary.
package transport

import (
	"context"
	"time"

	"connectrpc.com/connect"
)

// maxMessageBytes caps request and response sizes, matching the TS client's
// 100 MB limit so large attachments stream inline.
const maxMessageBytes = 100 << 20

// Config is the resolved transport configuration produced by the root package.
type Config struct {
	// Address is the server host:port.
	Address string
	// Insecure connects over cleartext HTTP/2 (h2c) instead of TLS.
	Insecure bool
	// Timeout is the default unary deadline; zero disables it.
	Timeout time.Duration
	// RetryEnabled turns on the unary retry interceptor.
	RetryEnabled bool
	// Retry parameterizes the backoff when RetryEnabled is set.
	Retry RetryConfig
	// AutoIdempotency attaches a generated key to mutating unary RPCs.
	AutoIdempotency bool
	// HTTPClient, when non-nil, overrides the constructed HTTP/2 client.
	HTTPClient connect.HTTPClient
	// Token resolves the bearer token for each request.
	Token func(ctx context.Context) (string, error)
}

// RetryConfig parameterizes the retry backoff.
type RetryConfig struct {
	MaxAttempts  int
	InitialDelay time.Duration
	MaxDelay     time.Duration
}
