package transport

import (
	"context"
	"errors"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
)

const (
	authorizationHeader = "Authorization"
	idempotencyHeader   = "X-Idempotency-Key"
	retryableHeader     = "X-Retryable"
)

// unaryInterceptor adapts a unary-only wrapper to a connect.Interceptor that
// passes streaming calls through unchanged. Retry, timeout, and idempotency use
// it so that, by construction, no streaming RPC is ever retried or deadlined.
type unaryInterceptor func(connect.UnaryFunc) connect.UnaryFunc

func (f unaryInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc { return f(next) }
func (f unaryInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}
func (f unaryInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return next
}

// authInterceptor injects the bearer token on both unary and streaming calls.
type authInterceptor struct {
	token func(ctx context.Context) (string, error)
}

func (a authInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		tok, err := a.token(ctx)
		if err != nil {
			return nil, connect.NewError(connect.CodeUnauthenticated, err)
		}
		req.Header().Set(authorizationHeader, "Bearer "+tok)
		return next(ctx, req)
	}
}

func (a authInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		conn := next(ctx, spec)
		// Headers must be set before the first Send flushes them. A token
		// error here can't be returned, so the server rejects the stream.
		if tok, err := a.token(ctx); err == nil {
			conn.RequestHeader().Set(authorizationHeader, "Bearer "+tok)
		}
		return conn
	}
}

func (a authInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return next
}

// timeoutInterceptor applies a default deadline to unary calls whose context
// has none.
func timeoutInterceptor(d time.Duration) unaryInterceptor {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			if _, ok := ctx.Deadline(); !ok {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, d)
				defer cancel()
			}
			return next(ctx, req)
		}
	}
}

// idempotencyInterceptor attaches a generated key to mutating unary RPCs. It
// runs once per logical call (outside retry), so retried attempts reuse the key.
func idempotencyInterceptor(newKey func() string) unaryInterceptor {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			if _, ok := mutatingProcedures[req.Spec().Procedure]; ok {
				if req.Header().Get(idempotencyHeader) == "" {
					req.Header().Set(idempotencyHeader, newKey())
				}
			}
			return next(ctx, req)
		}
	}
}

// retryInterceptor retries server-marked-retryable unary failures with
// exponential backoff and full jitter. rng returns a value in [0, 1).
func retryInterceptor(cfg RetryConfig, rng func() float64) unaryInterceptor {
	if cfg.MaxAttempts < 1 {
		cfg.MaxAttempts = 1
	}
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			var lastErr error
			for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
				resp, err := next(ctx, req)
				if err == nil {
					return resp, nil
				}
				lastErr = err
				if attempt == cfg.MaxAttempts-1 || !serverRetryable(err) {
					return nil, err
				}
				timer := time.NewTimer(backoff(cfg, attempt, rng))
				select {
				case <-ctx.Done():
					timer.Stop()
					return nil, ctx.Err()
				case <-timer.C:
				}
			}
			return nil, lastErr
		}
	}
}

// serverRetryable reports whether the server tagged the failure as retryable.
func serverRetryable(err error) bool {
	var ce *connect.Error
	if errors.As(err, &ce) {
		return ce.Meta().Get(retryableHeader) == "true"
	}
	return false
}

// backoff returns initialDelay*2^attempt capped at maxDelay, then applies full
// jitter in [0, cap).
func backoff(cfg RetryConfig, attempt int, rng func() float64) time.Duration {
	d := cfg.InitialDelay << attempt
	if d <= 0 || d > cfg.MaxDelay {
		d = cfg.MaxDelay
	}
	return time.Duration(rng() * float64(d))
}

func newIdempotencyKey() string { return uuid.NewString() }

// mutatingProcedures is the set of state-changing unary RPCs that receive an
// idempotency key. Mirrors the TS client's MUTATING_METHODS set.
var mutatingProcedures = map[string]struct{}{
	// MessageService
	"/photon.imessage.v1.MessageService/SendTextMessage":              {},
	"/photon.imessage.v1.MessageService/SendAttachmentMessage":        {},
	"/photon.imessage.v1.MessageService/SendMultipartMessage":         {},
	"/photon.imessage.v1.MessageService/SendMiniAppMessage":           {},
	"/photon.imessage.v1.MessageService/SendCustomizedMiniAppMessage": {},
	"/photon.imessage.v1.MessageService/EditMessage":                  {},
	"/photon.imessage.v1.MessageService/UnsendMessage":                {},
	"/photon.imessage.v1.MessageService/SetReaction":                  {},
	"/photon.imessage.v1.MessageService/PlaceSticker":                 {},
	"/photon.imessage.v1.MessageService/NotifySilencedMessage":        {},
	// ChatService
	"/photon.imessage.v1.ChatService/CreateChat":       {},
	"/photon.imessage.v1.ChatService/MarkChatRead":     {},
	"/photon.imessage.v1.ChatService/SetBackground":    {},
	"/photon.imessage.v1.ChatService/RemoveBackground": {},
	"/photon.imessage.v1.ChatService/ShareContactInfo": {},
	"/photon.imessage.v1.ChatService/SetTyping":        {},
	// GroupService
	"/photon.imessage.v1.GroupService/SetDisplayName":     {},
	"/photon.imessage.v1.GroupService/AddParticipants":    {},
	"/photon.imessage.v1.GroupService/RemoveParticipants": {},
	"/photon.imessage.v1.GroupService/LeaveGroup":         {},
	"/photon.imessage.v1.GroupService/SetIcon":            {},
	"/photon.imessage.v1.GroupService/RemoveIcon":         {},
	// AttachmentService
	"/photon.imessage.v1.AttachmentService/UploadAttachment": {},
	// PollService
	"/photon.imessage.v1.PollService/CreatePoll":    {},
	"/photon.imessage.v1.PollService/VotePoll":      {},
	"/photon.imessage.v1.PollService/UnvotePoll":    {},
	"/photon.imessage.v1.PollService/AddPollOption": {},
	// LocationService
	"/photon.imessage.v1.LocationService/RequestFriendLocationSharing": {},
}
