package imessage

import "context"

// A TokenSource supplies bearer tokens used to authenticate every request. It
// is the Advanced iMessage analogue of [golang.org/x/oauth2.TokenSource].
//
// Token is called once per RPC. Implementations that fetch rotating credentials
// should cache and refresh internally, and must be safe for concurrent use.
type TokenSource interface {
	// Token returns the bearer token to present, or an error to fail the RPC
	// before it is sent.
	Token(ctx context.Context) (string, error)
}

// StaticToken returns a [TokenSource] that always yields the same token. Use it
// for a long-lived API key:
//
//	client, err := imessage.New(addr, imessage.StaticToken(apiKey))
func StaticToken(token string) TokenSource { return staticToken(token) }

type staticToken string

func (t staticToken) Token(context.Context) (string, error) { return string(t), nil }

// TokenFunc adapts an ordinary function to a [TokenSource]. Use it for rotating
// credentials:
//
//	src := imessage.TokenFunc(func(ctx context.Context) (string, error) {
//		return vault.FetchToken(ctx)
//	})
type TokenFunc func(ctx context.Context) (string, error)

// Token calls f.
func (f TokenFunc) Token(ctx context.Context) (string, error) { return f(ctx) }

// Compile-time checks that the helpers satisfy TokenSource.
var (
	_ TokenSource = staticToken("")
	_ TokenSource = TokenFunc(nil)
)
