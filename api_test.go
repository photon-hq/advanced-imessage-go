package imessage_test

import (
	"context"
	"slices"
	"sync/atomic"
	"testing"

	imessagev1connect "buf.build/gen/go/photon-hq/imessage/connectrpc/go/photon/imessage/v1/imessagev1connect"
	imessagev1 "buf.build/gen/go/photon-hq/imessage/protocolbuffers/go/photon/imessage/v1"
	"connectrpc.com/connect"

	imessage "github.com/photon-hq/advanced-imessage-go"
)

func TestNewMessageSubscriptionYieldsCannedEvents(t *testing.T) {
	want := []imessage.MessageEvent{
		imessage.MessageReceived{Message: imessage.Message{GUID: "m1"}},
		imessage.MessageReactionAdded{MessageGUID: "m2"},
	}
	sub := imessage.NewMessageSubscription(slices.Values(want))
	defer sub.Close()

	var got []imessage.MessageEvent
	for ev := range sub.Events() {
		got = append(got, ev)
	}
	if err := sub.Err(); err != nil {
		t.Fatalf("Err() = %v, want nil", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d events, want 2", len(got))
	}
	if _, ok := got[0].(imessage.MessageReceived); !ok {
		t.Errorf("event 0 type = %T, want MessageReceived", got[0])
	}
	if _, ok := got[1].(imessage.MessageReactionAdded); !ok {
		t.Errorf("event 1 type = %T, want MessageReactionAdded", got[1])
	}
	// Close is idempotent.
	if err := sub.Close(); err != nil {
		t.Errorf("second Close() = %v", err)
	}
}

func TestNewSubscriptionNilIteratorIsEmpty(t *testing.T) {
	sub := imessage.NewChatSubscription(nil)
	defer sub.Close()
	for range sub.Events() {
		t.Fatal("nil iterator yielded an event")
	}
	if err := sub.Err(); err != nil {
		t.Errorf("Err() = %v, want nil", err)
	}
}

// fakeReaction is a MessageService that only implements SetReaction.
type fakeReaction struct {
	imessagev1connect.UnimplementedMessageServiceHandler
	setReaction func(context.Context, *connect.Request[imessagev1.SetReactionRequest]) (*connect.Response[imessagev1.MessageResponse], error)
}

func (f fakeReaction) SetReaction(ctx context.Context, req *connect.Request[imessagev1.SetReactionRequest]) (*connect.Response[imessagev1.MessageResponse], error) {
	return f.setReaction(ctx, req)
}

func TestAddReactionValidatesKind(t *testing.T) {
	var called atomic.Int32
	fake := fakeReaction{setReaction: func(context.Context, *connect.Request[imessagev1.SetReactionRequest]) (*connect.Response[imessagev1.MessageResponse], error) {
		called.Add(1)
		msg := &imessagev1.Message{}
		msg.SetGuid("reacted")
		resp := &imessagev1.MessageResponse{}
		resp.SetMessage(msg)
		return connect.NewResponse(resp), nil
	}}
	c := newClient(t, mountMessage(fake))
	ctx := context.Background()
	const chat = imessage.ChatGUID("iMessage;-;+1")

	// Non-settable / malformed kinds are rejected client-side, before any RPC.
	bad := []imessage.Reaction{
		{Kind: imessage.ReactionSticker},
		{Kind: imessage.ReactionUnknown},
		{Kind: imessage.MessageReactionKind("bogus")},
		{Kind: imessage.ReactionEmoji}, // emoji kind with no emoji
	}
	for _, r := range bad {
		if _, err := c.Messages().AddReaction(ctx, chat, "m1", r, nil); imessage.CodeOf(err) != imessage.CodeInvalidArgument {
			t.Errorf("AddReaction(%q) code = %q, want %q", r.Kind, imessage.CodeOf(err), imessage.CodeInvalidArgument)
		}
	}
	if n := called.Load(); n != 0 {
		t.Errorf("server SetReaction called %d times for invalid input, want 0", n)
	}

	// Valid reactions reach the server (add and remove).
	if _, err := c.Messages().AddReaction(ctx, chat, "m1", imessage.Reaction{Kind: imessage.ReactionLove}, nil); err != nil {
		t.Fatalf("AddReaction(love): %v", err)
	}
	if _, err := c.Messages().RemoveReaction(ctx, chat, "m1", imessage.Reaction{Kind: imessage.ReactionEmoji, Emoji: "🎉"}, nil); err != nil {
		t.Fatalf("RemoveReaction(emoji): %v", err)
	}
	if n := called.Load(); n != 2 {
		t.Errorf("server SetReaction called %d times for valid input, want 2", n)
	}
}
