package imessage_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	imessagev1connect "buf.build/gen/go/photon-hq/imessage/connectrpc/go/photon/imessage/v1/imessagev1connect"
	imessagev1 "buf.build/gen/go/photon-hq/imessage/protocolbuffers/go/photon/imessage/v1"
	"connectrpc.com/connect"
	"go.uber.org/goleak"

	imessage "github.com/photon-hq/advanced-imessage-go"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		// HTTP/2 client and server keep background goroutines alive between
		// requests; they are framework noise, not leaks in this library.
		goleak.IgnoreTopFunction("internal/poll.runtime_pollWait"),
		goleak.IgnoreTopFunction("net/http.(*persistConn).readLoop"),
		goleak.IgnoreTopFunction("net/http.(*persistConn).writeLoop"),
		goleak.IgnoreAnyFunction("golang.org/x/net/http2.(*ClientConn).readLoop"),
		goleak.IgnoreAnyFunction("golang.org/x/net/http2.(*serverConn).serve"),
		goleak.IgnoreAnyFunction("net/http.(*Server).Serve"),
	)
}

// newClient stands up an in-memory h2c connect server with the given mounted
// services and returns a Client wired to it.
func newClient(t *testing.T, mounts ...func(*http.ServeMux)) *imessage.Client {
	t.Helper()
	mux := http.NewServeMux()
	for _, mount := range mounts {
		mount(mux)
	}
	srv := httptest.NewUnstartedServer(mux)
	protocols := new(http.Protocols)
	protocols.SetUnencryptedHTTP2(true)
	srv.Config.Protocols = protocols
	srv.Start()
	t.Cleanup(srv.Close)

	addr := strings.TrimPrefix(srv.URL, "http://")
	c, err := imessage.New(addr, imessage.StaticToken("test-token"),
		imessage.WithoutTLS(),
		imessage.WithRetry(imessage.WithInitialDelay(time.Millisecond), imessage.WithMaxDelay(2*time.Millisecond)),
		imessage.WithAutoIdempotency(),
	)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return c
}

// --- fakes ---

type fakeAddress struct {
	imessagev1connect.UnimplementedAddressServiceHandler
	getInfo func(context.Context, *imessagev1.GetAddressInfoRequest) (*imessagev1.GetAddressInfoResponse, error)
}

func (f fakeAddress) GetAddressInfo(ctx context.Context, req *connect.Request[imessagev1.GetAddressInfoRequest]) (*connect.Response[imessagev1.GetAddressInfoResponse], error) {
	resp, err := f.getInfo(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func mountAddress(svc imessagev1connect.AddressServiceHandler) func(*http.ServeMux) {
	return func(mux *http.ServeMux) {
		path, h := imessagev1connect.NewAddressServiceHandler(svc)
		mux.Handle(path, h)
	}
}

type fakeMessage struct {
	imessagev1connect.UnimplementedMessageServiceHandler
	sendText  func(context.Context, *connect.Request[imessagev1.SendTextMessageRequest]) (*connect.Response[imessagev1.MessageResponse], error)
	subscribe func(context.Context, *connect.Request[imessagev1.SubscribeMessageEventsRequest], *connect.ServerStream[imessagev1.SubscribeMessageEventsResponse]) error
}

func (f fakeMessage) SendTextMessage(ctx context.Context, req *connect.Request[imessagev1.SendTextMessageRequest]) (*connect.Response[imessagev1.MessageResponse], error) {
	return f.sendText(ctx, req)
}

func (f fakeMessage) SubscribeMessageEvents(ctx context.Context, req *connect.Request[imessagev1.SubscribeMessageEventsRequest], stream *connect.ServerStream[imessagev1.SubscribeMessageEventsResponse]) error {
	return f.subscribe(ctx, req, stream)
}

func mountMessage(svc imessagev1connect.MessageServiceHandler) func(*http.ServeMux) {
	return func(mux *http.ServeMux) {
		path, h := imessagev1connect.NewMessageServiceHandler(svc)
		mux.Handle(path, h)
	}
}

// --- tests ---

func TestAddressGet(t *testing.T) {
	fake := fakeAddress{getInfo: func(_ context.Context, req *imessagev1.GetAddressInfoRequest) (*imessagev1.GetAddressInfoResponse, error) {
		if req.GetAddress() != "+15551234567" {
			t.Errorf("server saw address %q", req.GetAddress())
		}
		info := &imessagev1.MultiServiceAddressInfo{}
		info.SetAddress("+15551234567")
		info.SetServices([]imessagev1.ChatServiceType{
			imessagev1.ChatServiceType_CHAT_SERVICE_TYPE_IMESSAGE,
			imessagev1.ChatServiceType_CHAT_SERVICE_TYPE_SMS,
		})
		info.SetCountry("us")
		resp := &imessagev1.GetAddressInfoResponse{}
		resp.SetInfo(info)
		return resp, nil
	}}
	c := newClient(t, mountAddress(fake))

	got, err := c.Addresses().Get(context.Background(), "+15551234567")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Address != "+15551234567" || got.Country != "us" {
		t.Errorf("got %+v", got)
	}
	if len(got.Services) != 2 || got.Services[0] != imessage.ServiceIMessage || got.Services[1] != imessage.ServiceSMS {
		t.Errorf("services = %v", got.Services)
	}
}

func TestErrorTranslation(t *testing.T) {
	fake := fakeAddress{getInfo: func(context.Context, *imessagev1.GetAddressInfoRequest) (*imessagev1.GetAddressInfoResponse, error) {
		cerr := connect.NewError(connect.CodeNotFound, errors.New("not seen"))
		cerr.Meta().Set("error-code", string(imessage.CodeAddressNotFound))
		cerr.Meta().Set("x-retryable", "false")
		return nil, cerr
	}}
	c := newClient(t, mountAddress(fake))

	_, err := c.Addresses().Get(context.Background(), "+1")
	if err == nil {
		t.Fatal("expected error")
	}
	if !imessage.IsNotFound(err) {
		t.Errorf("IsNotFound = false for %v", err)
	}
	if imessage.CodeOf(err) != imessage.CodeAddressNotFound {
		t.Errorf("CodeOf = %q", imessage.CodeOf(err))
	}
}

func TestSendTextMappingAndIdempotency(t *testing.T) {
	var gotKey, gotChat, gotText string
	fake := fakeMessage{sendText: func(_ context.Context, req *connect.Request[imessagev1.SendTextMessageRequest]) (*connect.Response[imessagev1.MessageResponse], error) {
		gotKey = req.Header().Get("X-Idempotency-Key")
		gotChat = req.Msg.GetChatGuid()
		gotText = req.Msg.GetText()
		msg := &imessagev1.Message{}
		msg.SetGuid("sent-1")
		content := &imessagev1.MessageContent{}
		content.SetText("hello")
		msg.SetContent(content)
		resp := &imessagev1.MessageResponse{}
		resp.SetMessage(msg)
		return connect.NewResponse(resp), nil
	}}
	c := newClient(t, mountMessage(fake))

	msg, err := c.Messages().SendText(context.Background(), "iMessage;-;+15551234567", "hello", nil)
	if err != nil {
		t.Fatalf("SendText: %v", err)
	}
	if msg.GUID != "sent-1" || msg.Content.Text != "hello" {
		t.Errorf("mapped message = %+v", msg)
	}
	if gotChat != "iMessage;-;+15551234567" || gotText != "hello" {
		t.Errorf("server saw chat=%q text=%q", gotChat, gotText)
	}
	if gotKey == "" {
		t.Error("expected an idempotency key header on the mutating call")
	}
}

func TestRetryHonorsServerSignal(t *testing.T) {
	var calls atomic.Int32
	fake := fakeAddress{getInfo: func(context.Context, *imessagev1.GetAddressInfoRequest) (*imessagev1.GetAddressInfoResponse, error) {
		if calls.Add(1) < 3 {
			cerr := connect.NewError(connect.CodeUnavailable, errors.New("try later"))
			cerr.Meta().Set("x-retryable", "true")
			return nil, cerr
		}
		info := &imessagev1.MultiServiceAddressInfo{}
		info.SetAddress("+1")
		resp := &imessagev1.GetAddressInfoResponse{}
		resp.SetInfo(info)
		return resp, nil
	}}
	c := newClient(t, mountAddress(fake))

	if _, err := c.Addresses().Get(context.Background(), "+1"); err != nil {
		t.Fatalf("Get after retries: %v", err)
	}
	if n := calls.Load(); n != 3 {
		t.Errorf("server call count = %d, want 3", n)
	}
}

func TestSubscribeDeliversEventsAndDropsHeartbeat(t *testing.T) {
	fake := fakeMessage{subscribe: func(_ context.Context, _ *connect.Request[imessagev1.SubscribeMessageEventsRequest], stream *connect.ServerStream[imessagev1.SubscribeMessageEventsResponse]) error {
		hb := &imessagev1.SubscribeMessageEventsResponse{}
		hb.SetHeartbeat(&imessagev1.Heartbeat{})
		if err := stream.Send(hb); err != nil {
			return err
		}
		for i := 0; i < 2; i++ {
			msg := &imessagev1.Message{}
			msg.SetGuid(fmt.Sprintf("m%d", i))
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
		return nil
	}}
	c := newClient(t, mountMessage(fake))

	sub := c.Messages().Subscribe(context.Background(), nil)
	defer sub.Close()

	var guids []string
	for ev := range sub.Events() {
		recv, ok := ev.(imessage.MessageReceived)
		if !ok {
			t.Fatalf("unexpected event type %T", ev)
		}
		guids = append(guids, recv.Message.GUID)
	}
	if err := sub.Err(); err != nil {
		t.Fatalf("stream Err: %v", err)
	}
	if len(guids) != 2 || guids[0] != "m0" || guids[1] != "m1" {
		t.Errorf("received guids = %v", guids)
	}
}

func TestSubscribeEarlyCloseNoLeak(t *testing.T) {
	// An endless stream; the consumer breaks after one event and Closes. The
	// goleak check in TestMain verifies the pump goroutine is reclaimed.
	fake := fakeMessage{subscribe: func(ctx context.Context, _ *connect.Request[imessagev1.SubscribeMessageEventsRequest], stream *connect.ServerStream[imessagev1.SubscribeMessageEventsResponse]) error {
		for {
			msg := &imessagev1.Message{}
			msg.SetGuid("loop")
			recv := &imessagev1.MessageReceived{}
			recv.SetMessage(msg)
			change := &imessagev1.MessageChangeEvent{}
			change.SetMessageReceived(recv)
			resp := &imessagev1.SubscribeMessageEventsResponse{}
			resp.SetMessageChanged(change)
			if err := stream.Send(resp); err != nil {
				return err
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Millisecond):
			}
		}
	}}
	c := newClient(t, mountMessage(fake))

	sub := c.Messages().Subscribe(context.Background(), nil)
	for range sub.Events() {
		break // stop after the first event
	}
	if err := sub.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}
