package imessage

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
)

func TestAsErrorMapsConnectError(t *testing.T) {
	ce := connect.NewError(connect.CodeNotFound, errors.New("missing"))
	ce.Meta().Set("error-code", string(CodeChatNotFound))
	ce.Meta().Set("x-retryable", "true")
	ce.Meta().Set("error-context-chat", "iMessage;-;x")

	err := asError(ce)

	var ie *Error
	if !errors.As(err, &ie) {
		t.Fatalf("asError returned %T, want *Error", err)
	}
	if ie.Code != CodeChatNotFound {
		t.Errorf("Code = %q, want %q", ie.Code, CodeChatNotFound)
	}
	if ie.ConnectCode != connect.CodeNotFound {
		t.Errorf("ConnectCode = %v, want NotFound", ie.ConnectCode)
	}
	if !ie.Retryable {
		t.Error("Retryable = false, want true")
	}
	if got := ie.Context["chat"]; got != "iMessage;-;x" {
		t.Errorf("Context[chat] = %q, want %q", got, "iMessage;-;x")
	}
	if !IsNotFound(err) {
		t.Error("IsNotFound = false")
	}
	if !IsRetryable(err) {
		t.Error("IsRetryable = false")
	}
	if CodeOf(err) != CodeChatNotFound {
		t.Errorf("CodeOf = %q", CodeOf(err))
	}
}

func TestErrorIsByClass(t *testing.T) {
	rateLimited := &Error{ConnectCode: connect.CodeResourceExhausted}
	if !errors.Is(rateLimited, ErrRateLimited) {
		t.Error("ResourceExhausted should match ErrRateLimited")
	}
	if !IsRateLimited(rateLimited) {
		t.Error("IsRateLimited = false")
	}
	if errors.Is(rateLimited, ErrNotFound) {
		t.Error("ResourceExhausted should not match ErrNotFound")
	}

	specific := &Error{Code: CodeChatNotFound, ConnectCode: connect.CodeNotFound}
	if !errors.Is(specific, &Error{Code: CodeChatNotFound}) {
		t.Error("specific code should match by code")
	}
	if errors.Is(specific, &Error{Code: CodeMessageNotFound}) {
		t.Error("different code should not match")
	}
}

func TestAsErrorPassThroughAndNil(t *testing.T) {
	if asError(nil) != nil {
		t.Error("asError(nil) should be nil")
	}
	sentinel := errors.New("library error")
	if got := asError(sentinel); got != sentinel { //nolint:errorlint // identity check intended
		t.Errorf("asError should pass non-RPC errors through, got %v", got)
	}
}

func TestAsErrorMapsContextDeadline(t *testing.T) {
	err := asError(context.DeadlineExceeded)
	if !IsConnection(err) {
		t.Error("DeadlineExceeded should map to a connection-class error")
	}
	if CodeOf(err) != CodeTimeout {
		t.Errorf("CodeOf = %q, want %q", CodeOf(err), CodeTimeout)
	}
}
