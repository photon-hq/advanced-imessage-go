package imessage

import (
	"time"

	"connectrpc.com/connect"
)

// Default retry parameters, matching the TypeScript client.
const (
	defaultMaxAttempts  = 4
	defaultInitialDelay = 200 * time.Millisecond
	defaultMaxDelay     = 5 * time.Second
)

// An Option configures a [Client] during [New]. Options are applied in order.
type Option interface{ apply(*config) }

type optionFunc func(*config)

func (f optionFunc) apply(c *config) { f(c) }

// config holds the resolved client configuration.
type config struct {
	insecure        bool
	timeout         time.Duration
	retryEnabled    bool
	retry           RetryOptions
	autoIdempotency bool
	httpClient      connect.HTTPClient
}

func defaultConfig() config {
	return config{
		retry: RetryOptions{
			MaxAttempts:  defaultMaxAttempts,
			InitialDelay: defaultInitialDelay,
			MaxDelay:     defaultMaxDelay,
		},
	}
}

// WithoutTLS disables TLS, connecting over cleartext HTTP/2 (h2c). TLS is on by
// default; use this only for local development against an insecure endpoint.
func WithoutTLS() Option {
	return optionFunc(func(c *config) { c.insecure = true })
}

// WithTimeout sets a default deadline applied to unary RPCs whose context does
// not already carry one. Streaming RPCs are never subject to it. A zero or
// negative duration disables the default (the inherited context still applies).
func WithTimeout(d time.Duration) Option {
	return optionFunc(func(c *config) { c.timeout = d })
}

// WithRetry enables automatic retry of unary RPCs that the server marks as
// retryable, using exponential backoff with full jitter. Without any
// [RetryOption] it uses the package defaults (4 attempts, 200ms initial, 5s
// max). Streaming RPCs are never retried.
func WithRetry(opts ...RetryOption) Option {
	return optionFunc(func(c *config) {
		c.retryEnabled = true
		if c.retry.MaxAttempts == 0 {
			c.retry = RetryOptions{
				MaxAttempts:  defaultMaxAttempts,
				InitialDelay: defaultInitialDelay,
				MaxDelay:     defaultMaxDelay,
			}
		}
		for _, o := range opts {
			o(&c.retry)
		}
	})
}

// WithAutoIdempotency attaches a generated idempotency key to every mutating
// RPC so the server can de-duplicate retried writes. Calls that already set a
// client message id keep their own key.
func WithAutoIdempotency() Option {
	return optionFunc(func(c *config) { c.autoIdempotency = true })
}

// WithHTTPClient overrides the underlying HTTP client. This is an advanced
// option; the supplied client must support HTTP/2. When set, [WithoutTLS] has
// no effect on transport construction (the caller owns the transport).
func WithHTTPClient(hc connect.HTTPClient) Option {
	return optionFunc(func(c *config) { c.httpClient = hc })
}

// RetryOptions parameterizes the retry backoff. See [WithRetry].
type RetryOptions struct {
	// MaxAttempts is the total number of attempts, including the first.
	MaxAttempts int
	// InitialDelay is the base backoff before the second attempt.
	InitialDelay time.Duration
	// MaxDelay caps the per-attempt backoff.
	MaxDelay time.Duration
}

// A RetryOption tunes [RetryOptions] for [WithRetry].
type RetryOption func(*RetryOptions)

// WithMaxAttempts sets the total number of attempts (including the first).
func WithMaxAttempts(n int) RetryOption {
	return func(r *RetryOptions) { r.MaxAttempts = n }
}

// WithInitialDelay sets the base backoff before the second attempt.
func WithInitialDelay(d time.Duration) RetryOption {
	return func(r *RetryOptions) { r.InitialDelay = d }
}

// WithMaxDelay caps the per-attempt backoff.
func WithMaxDelay(d time.Duration) RetryOption {
	return func(r *RetryOptions) { r.MaxDelay = d }
}
