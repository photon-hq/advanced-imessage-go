package imessage

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"connectrpc.com/connect"
)

// Code is a canonical, machine-readable failure code returned by the Advanced
// iMessage backend. It is carried on every [*Error] and mirrors the string
// codes used by the TypeScript client.
type Code string

// Canonical error codes. The zero value ("") means the server did not report a
// specific code; callers should branch on the category sentinels (for example
// [ErrNotFound]) rather than on individual codes whenever possible.
const (
	// Authentication / authorization.
	CodeUnauthenticated Code = "unauthenticated"
	CodeTokenExpired    Code = "tokenExpired"
	CodeTokenBlocked    Code = "tokenBlocked"
	CodeUnauthorized    Code = "unauthorized"

	// Rate limiting.
	CodeDailyLimitExceeded       Code = "dailyLimitExceeded"
	CodeRecipientLimitExceeded   Code = "recipientLimitExceeded"
	CodeUploadRateExceeded       Code = "uploadRateExceeded"
	CodeContentDuplicateExceeded Code = "contentDuplicateExceeded"
	CodeRecipientCoolingDown     Code = "recipientCoolingDown"
	CodeRecipientLocked          Code = "recipientLocked"
	CodeSendReceiveRatioExceeded Code = "sendReceiveRatioExceeded"

	// Duplicate.
	CodeDuplicateMessage Code = "duplicateMessage"

	// Not found.
	CodeChatNotFound                 Code = "chatNotFound"
	CodeMessageNotFound              Code = "messageNotFound"
	CodeAttachmentNotFound           Code = "attachmentNotFound"
	CodeAddressNotFound              Code = "addressNotFound"
	CodeSharedFriendLocationNotFound Code = "sharedFriendLocationNotFound"
	CodeGroupIconNotFound            Code = "groupIconNotFound"
	CodePollNotFound                 Code = "pollNotFound"

	// Validation / precondition.
	CodeInvalidArgument       Code = "invalidArgument"
	CodePreconditionFailed    Code = "preconditionFailed"
	CodeOperationNotSupported Code = "operationNotSupported"
	CodeAttachmentNotReady    Code = "attachmentNotReady"
	CodePrivateAPIUnavailable Code = "privateApiUnavailable"

	// Infrastructure.
	CodeServiceUnavailable Code = "serviceUnavailable"
	CodeTimeout            Code = "timeout"
	CodeInternalError      Code = "internalError"
	CodeDatabaseError      Code = "databaseError"
	CodeNetworkError       Code = "networkError"
)

// Error is the single error type returned by every Client operation. Match its
// category with [errors.Is] against the package sentinels (such as
// [ErrNotFound]) or the Is* helpers, and read its structured detail with
// [errors.As]:
//
//	var ierr *imessage.Error
//	if errors.As(err, &ierr) {
//		log.Printf("code=%s retryable=%v", ierr.Code, ierr.Retryable)
//	}
type Error struct {
	// Code is the canonical server code, or "" if none was reported.
	Code Code
	// ConnectCode is the underlying Connect/gRPC status code.
	ConnectCode connect.Code
	// Message is a human-readable description from the server.
	Message string
	// Retryable reports whether the server marked the failure as safe to retry.
	Retryable bool
	// Context carries any structured key/value detail attached by the server.
	Context map[string]string

	cause error
}

// Error implements the error interface.
func (e *Error) Error() string {
	var b strings.Builder
	b.WriteString("imessage: ")
	if e.Code != "" {
		b.WriteString(string(e.Code))
		b.WriteString(": ")
	}
	if e.Message != "" {
		b.WriteString(e.Message)
	} else {
		b.WriteString(e.ConnectCode.String())
	}
	return b.String()
}

// Unwrap returns the underlying transport error, if any.
func (e *Error) Unwrap() error { return e.cause }

// Is reports whether target matches this error. A target carrying a specific
// [Code] matches by code; a category sentinel matches by status class. This is
// what lets errors.Is(err, ErrNotFound) and
// errors.Is(err, &imessage.Error{Code: imessage.CodeChatNotFound}) both work.
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return false
	}
	if t.Code != "" {
		return e.Code == t.Code
	}
	return classOf(e.ConnectCode) == classOf(t.ConnectCode)
}

// errClass groups Connect status codes into the user-facing categories exposed
// as sentinels.
type errClass int

const (
	classInternal errClass = iota
	classAuth
	classNotFound
	classRateLimit
	classValidation
	classConnection
)

func classOf(c connect.Code) errClass {
	switch c {
	case connect.CodeUnauthenticated, connect.CodePermissionDenied:
		return classAuth
	case connect.CodeNotFound:
		return classNotFound
	case connect.CodeResourceExhausted:
		return classRateLimit
	case connect.CodeInvalidArgument, connect.CodeFailedPrecondition:
		return classValidation
	case connect.CodeUnavailable, connect.CodeDeadlineExceeded, connect.CodeCanceled:
		return classConnection
	default:
		return classInternal
	}
}

// Category sentinels for errors.Is. Each carries only the representative
// Connect status code for its class; matching is by class, not exact code.
var (
	// ErrAuthentication matches authentication and permission failures.
	ErrAuthentication = &Error{ConnectCode: connect.CodeUnauthenticated}
	// ErrNotFound matches "resource does not exist" failures.
	ErrNotFound = &Error{ConnectCode: connect.CodeNotFound}
	// ErrRateLimited matches rate-limit / quota failures.
	ErrRateLimited = &Error{ConnectCode: connect.CodeResourceExhausted}
	// ErrValidation matches invalid-argument / failed-precondition failures.
	ErrValidation = &Error{ConnectCode: connect.CodeInvalidArgument}
	// ErrConnection matches transport / availability / deadline failures.
	ErrConnection = &Error{ConnectCode: connect.CodeUnavailable}
)

// IsAuthentication reports whether err is an authentication or permission error.
func IsAuthentication(err error) bool { return errors.Is(err, ErrAuthentication) }

// IsNotFound reports whether err indicates a missing resource.
func IsNotFound(err error) bool { return errors.Is(err, ErrNotFound) }

// IsRateLimited reports whether err is a rate-limit or quota error.
func IsRateLimited(err error) bool { return errors.Is(err, ErrRateLimited) }

// IsValidation reports whether err is an invalid-argument or precondition error.
func IsValidation(err error) bool { return errors.Is(err, ErrValidation) }

// IsConnection reports whether err is a transport, availability, or deadline error.
func IsConnection(err error) bool { return errors.Is(err, ErrConnection) }

// IsRetryable reports whether the server marked err as safe to retry.
func IsRetryable(err error) bool {
	var e *Error
	return errors.As(err, &e) && e.Retryable
}

// CodeOf returns the canonical [Code] carried by err, or "" if err is not an
// [*Error] or carried no code.
func CodeOf(err error) Code {
	var e *Error
	if errors.As(err, &e) {
		return e.Code
	}
	return ""
}

// asError converts a Connect/transport error into an [*Error]. It returns nil
// for a nil input and is idempotent for values that are already [*Error].
func asError(err error) error {
	if err == nil {
		return nil
	}
	var existing *Error
	if errors.As(err, &existing) {
		return existing
	}

	var ce *connect.Error
	if !errors.As(err, &ce) {
		// Context cancellation/deadline can reach us raw (e.g. from the retry
		// interceptor). Map them to connection-class errors so IsConnection
		// matches; pass anything else (such as library sentinels) through
		// unchanged rather than mislabeling it.
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			return &Error{Code: CodeTimeout, ConnectCode: connect.CodeDeadlineExceeded, Message: err.Error(), cause: err}
		case errors.Is(err, context.Canceled):
			return &Error{Code: CodeNetworkError, ConnectCode: connect.CodeCanceled, Message: err.Error(), cause: err}
		default:
			return err
		}
	}

	meta := ce.Meta()
	code := Code(firstMeta(meta, "error-code"))
	if code == "" {
		code = defaultCodeFor(ce.Code())
	}
	return &Error{
		Code:        code,
		ConnectCode: ce.Code(),
		Message:     ce.Message(),
		Retryable:   firstMeta(meta, "x-retryable") == "true",
		Context:     contextFromMeta(meta),
		cause:       err,
	}
}

// defaultCodeFor maps a Connect status code to a representative [Code] when the
// server did not send an explicit one.
func defaultCodeFor(c connect.Code) Code {
	switch c {
	case connect.CodeUnauthenticated:
		return CodeUnauthenticated
	case connect.CodePermissionDenied:
		return CodeUnauthorized
	case connect.CodeNotFound, connect.CodeResourceExhausted:
		// Ambiguous: the specific resource/limit is unknown without the
		// server's error-code header. Leave Code unset and let the category
		// helpers (IsNotFound, IsRateLimited) classify by status class.
		return ""
	case connect.CodeInvalidArgument:
		return CodeInvalidArgument
	case connect.CodeFailedPrecondition:
		return CodePreconditionFailed
	case connect.CodeUnimplemented:
		return CodeOperationNotSupported
	case connect.CodeUnavailable:
		return CodeServiceUnavailable
	case connect.CodeDeadlineExceeded:
		return CodeTimeout
	default:
		return CodeInternalError
	}
}

func firstMeta(h http.Header, key string) string {
	return h.Get(key)
}

func contextFromMeta(h http.Header) map[string]string {
	const prefix = "error-context-"
	var out map[string]string
	for k, vs := range h {
		lk := strings.ToLower(k)
		if !strings.HasPrefix(lk, prefix) || len(vs) == 0 {
			continue
		}
		if out == nil {
			out = make(map[string]string)
		}
		out[strings.TrimPrefix(lk, prefix)] = vs[0]
	}
	return out
}
