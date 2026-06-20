package imessage_test

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	imessagev1connect "buf.build/gen/go/photon-hq/imessage/connectrpc/go/photon/imessage/v1/imessagev1connect"
	imessagev1 "buf.build/gen/go/photon-hq/imessage/protocolbuffers/go/photon/imessage/v1"
	"connectrpc.com/connect"
)

type memStore struct {
	mu  sync.Mutex
	seq uint64
	ok  bool
}

func (s *memStore) Load(context.Context) (uint64, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.seq, s.ok, nil
}

func (s *memStore) Save(_ context.Context, seq uint64) error {
	s.mu.Lock()
	s.seq, s.ok = seq, true
	s.mu.Unlock()
	return nil
}

func (s *memStore) current() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.seq
}

type fakeEvent struct {
	imessagev1connect.UnimplementedEventServiceHandler
	catchUp func(context.Context, *connect.Request[imessagev1.CatchUpEventsRequest], *connect.ServerStream[imessagev1.CatchUpEventsResponse]) error
}

func (f fakeEvent) CatchUpEvents(ctx context.Context, req *connect.Request[imessagev1.CatchUpEventsRequest], stream *connect.ServerStream[imessagev1.CatchUpEventsResponse]) error {
	return f.catchUp(ctx, req, stream)
}

func mountEvent(svc imessagev1connect.EventServiceHandler) func(*http.ServeMux) {
	return func(mux *http.ServeMux) {
		path, h := imessagev1connect.NewEventServiceHandler(svc)
		mux.Handle(path, h)
	}
}

func TestResumableMessages(t *testing.T) {
	fe := fakeEvent{catchUp: func(_ context.Context, _ *connect.Request[imessagev1.CatchUpEventsRequest], stream *connect.ServerStream[imessagev1.CatchUpEventsResponse]) error {
		// No history; immediately report the head and complete.
		resp := &imessagev1.CatchUpEventsResponse{}
		complete := &imessagev1.CatchUpEventsComplete{}
		complete.SetHeadSequence(0)
		resp.SetComplete(complete)
		return stream.Send(resp)
	}}

	fm := fakeMessage{subscribe: func(ctx context.Context, _ *connect.Request[imessagev1.SubscribeMessageEventsRequest], stream *connect.ServerStream[imessagev1.SubscribeMessageEventsResponse]) error {
		for i := 0; i < 2; i++ {
			msg := &imessagev1.Message{}
			msg.SetGuid("live")
			recv := &imessagev1.MessageReceived{}
			recv.SetMessage(msg)
			change := &imessagev1.MessageChangeEvent{}
			change.SetChatGuid("iMessage;-;+1")
			change.SetMessageReceived(recv)
			resp := &imessagev1.SubscribeMessageEventsResponse{}
			resp.SetSequence(uint64(i + 1))
			resp.SetMessageChanged(change)
			if err := stream.Send(resp); err != nil {
				return err
			}
		}
		<-ctx.Done() // stay live until the client disconnects
		return ctx.Err()
	}}

	c := newClient(t, mountEvent(fe), mountMessage(fm))
	store := &memStore{}

	sub := c.ResumableMessages(context.Background(), store, nil)
	var count int
	for range sub.Events() {
		count++
		if count == 2 {
			break
		}
	}
	if err := sub.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	if count != 2 {
		t.Fatalf("received %d events, want 2", count)
	}
	if got := store.current(); got != 2 {
		t.Errorf("persisted sequence = %d, want 2", got)
	}
}

func TestResumableLoadsCursor(t *testing.T) {
	var sawAfter uint64
	fe := fakeEvent{catchUp: func(_ context.Context, req *connect.Request[imessagev1.CatchUpEventsRequest], stream *connect.ServerStream[imessagev1.CatchUpEventsResponse]) error {
		sawAfter = req.Msg.GetAfterSequence()
		resp := &imessagev1.CatchUpEventsResponse{}
		complete := &imessagev1.CatchUpEventsComplete{}
		complete.SetHeadSequence(sawAfter)
		resp.SetComplete(complete)
		return stream.Send(resp)
	}}
	fm := fakeMessage{subscribe: func(ctx context.Context, _ *connect.Request[imessagev1.SubscribeMessageEventsRequest], _ *connect.ServerStream[imessagev1.SubscribeMessageEventsResponse]) error {
		<-ctx.Done()
		return ctx.Err()
	}}

	c := newClient(t, mountEvent(fe), mountMessage(fm))
	store := &memStore{seq: 42, ok: true}

	sub := c.ResumableMessages(context.Background(), store, nil)
	// Give the engine a moment to run the initial catch-up, then stop.
	time.Sleep(50 * time.Millisecond)
	_ = sub.Close()

	if sawAfter != 42 {
		t.Errorf("catch-up after_sequence = %d, want 42 (loaded from store)", sawAfter)
	}
}
