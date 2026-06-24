package transport

import (
	"net/http"
	"testing"

	"golang.org/x/net/http2"
)

// http2TransportOf extracts the *http2.Transport that newHTTP2Client installs,
// failing the test if the client is not shaped as expected.
func http2TransportOf(t *testing.T, hc *http.Client) *http2.Transport {
	t.Helper()
	tr, ok := hc.Transport.(*http2.Transport)
	if !ok {
		t.Fatalf("transport = %T, want *http2.Transport", hc.Transport)
	}
	return tr
}

func TestNewHTTP2ClientSetsKeepalive(t *testing.T) {
	// Both the TLS and h2c paths must enable keepalive: a long-lived
	// server-streaming RPC over a half-open connection would otherwise block
	// the receive loop forever. ReadIdleTimeout drives the HTTP/2 PING and
	// PingTimeout bounds the wait for its reply.
	for _, insecure := range []bool{false, true} {
		tr := http2TransportOf(t, newHTTP2Client(insecure))
		if tr.ReadIdleTimeout != keepaliveTime {
			t.Errorf("insecure=%t: ReadIdleTimeout = %v, want %v", insecure, tr.ReadIdleTimeout, keepaliveTime)
		}
		if tr.PingTimeout != keepaliveTimeout {
			t.Errorf("insecure=%t: PingTimeout = %v, want %v", insecure, tr.PingTimeout, keepaliveTimeout)
		}
	}
}

func TestKeepaliveTimeNotBelowTimeout(t *testing.T) {
	// The ping interval must exceed the per-ping reply deadline, otherwise a
	// new health check would start before the previous one could be answered.
	if keepaliveTime <= keepaliveTimeout {
		t.Errorf("keepaliveTime (%v) must be greater than keepaliveTimeout (%v)", keepaliveTime, keepaliveTimeout)
	}
}
