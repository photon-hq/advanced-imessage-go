package transport

import (
	"context"
	"crypto/tls"
	"math/rand/v2"
	"net"
	"net/http"

	"connectrpc.com/connect"
	"golang.org/x/net/http2"
)

// Dial builds the HTTP client, base URL, and Connect client options (including
// the interceptor chain) from cfg. The returned options are passed to every
// generated service-client constructor so they share one connection pool and
// one interceptor stack.
func Dial(cfg Config) (httpClient connect.HTTPClient, baseURL string, opts []connect.ClientOption) {
	scheme := "https"
	if cfg.Insecure {
		scheme = "http"
	}
	baseURL = scheme + "://" + cfg.Address

	httpClient = cfg.HTTPClient
	if httpClient == nil {
		httpClient = newHTTP2Client(cfg.Insecure)
	}

	// Outer-to-inner: auth, timeout, idempotency, retry. Idempotency sits
	// outside retry so the same key is reused across attempts; timeout sits
	// outside retry so its deadline bounds the whole logical call.
	interceptors := []connect.Interceptor{authInterceptor{token: cfg.Token}}
	if cfg.Timeout > 0 {
		interceptors = append(interceptors, timeoutInterceptor(cfg.Timeout))
	}
	if cfg.AutoIdempotency {
		interceptors = append(interceptors, idempotencyInterceptor(newIdempotencyKey))
	}
	if cfg.RetryEnabled {
		interceptors = append(interceptors, retryInterceptor(cfg.Retry, rand.Float64))
	}

	opts = []connect.ClientOption{
		connect.WithGRPC(),
		connect.WithInterceptors(interceptors...),
		connect.WithReadMaxBytes(maxMessageBytes),
		connect.WithSendMaxBytes(maxMessageBytes),
	}
	return httpClient, baseURL, opts
}

// newHTTP2Client returns an HTTP/2 client. With TLS it relies on ALPN
// negotiation via the standard transport; insecure mode uses h2c (prior
// knowledge) by dialing cleartext TCP.
func newHTTP2Client(insecure bool) *http.Client {
	if !insecure {
		return &http.Client{Transport: &http2.Transport{}}
	}
	return &http.Client{
		Transport: &http2.Transport{
			AllowHTTP: true,
			DialTLSContext: func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
				var d net.Dialer
				return d.DialContext(ctx, network, addr)
			},
		},
	}
}
