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
	retry           retryConfig
	autoIdempotency bool
	httpClient      connect.HTTPClient
}

func defaultConfig() config {
	return config{
		retry: retryConfig{
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
			c.retry = retryConfig{
				MaxAttempts:  defaultMaxAttempts,
				InitialDelay: defaultInitialDelay,
				MaxDelay:     defaultMaxDelay,
			}
		}
		for _, o := range opts {
			o.applyRetry(&c.retry)
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

// retryConfig holds the resolved retry backoff parameters. Its fields mirror
// transport.RetryConfig so the root package can convert directly.
type retryConfig struct {
	MaxAttempts  int
	InitialDelay time.Duration
	MaxDelay     time.Duration
}

// A RetryOption tunes the retry backoff for [WithRetry]. The interface is closed
// (its only method is unexported), so the package fully controls the set of
// valid options; build them with [WithMaxAttempts], [WithInitialDelay], and
// [WithMaxDelay].
type RetryOption interface{ applyRetry(*retryConfig) }

type retryOptionFunc func(*retryConfig)

func (f retryOptionFunc) applyRetry(r *retryConfig) { f(r) }

// WithMaxAttempts sets the total number of attempts (including the first).
func WithMaxAttempts(n int) RetryOption {
	return retryOptionFunc(func(r *retryConfig) { r.MaxAttempts = n })
}

// WithInitialDelay sets the base backoff before the second attempt.
func WithInitialDelay(d time.Duration) RetryOption {
	return retryOptionFunc(func(r *retryConfig) { r.InitialDelay = d })
}

// WithMaxDelay caps the per-attempt backoff.
func WithMaxDelay(d time.Duration) RetryOption {
	return retryOptionFunc(func(r *retryConfig) { r.MaxDelay = d })
}

// Compile-time guards that the unexported adapters satisfy their interfaces.
var (
	_ Option      = optionFunc(nil)
	_ RetryOption = retryOptionFunc(nil)
)
