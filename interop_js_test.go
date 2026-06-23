//go:build interop

package imessage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"time"

	imessagev1connect "buf.build/gen/go/photon-hq/imessage/connectrpc/go/photon/imessage/v1/imessagev1connect"
	imessagev1 "buf.build/gen/go/photon-hq/imessage/protocolbuffers/go/photon/imessage/v1"
	"connectrpc.com/connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	interopChatGUID       = "any;-;alice@example.com"
	interopGroupChatGUID  = "any;+;group-guid"
	interopMessageGUID    = "message-guid"
	interopReplyGUID      = "reply-guid"
	interopAttachmentGUID = "attachment-guid"
	interopStickerGUID    = "sticker-guid"
	interopToken          = "interop-token"
)

var interopBaseTime = time.Date(2026, 6, 20, 12, 0, 0, 123_000_000, time.UTC)

func TestAdvancedIMessageTSInterop(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	scriptPath := filepath.Join("testdata", "interop-js", "client.mjs")
	if _, err := os.Stat(scriptPath); err != nil {
		t.Fatalf("missing JS interop client: %v", err)
	}
	packagePath := filepath.Join("testdata", "interop-js", "node_modules", "@photon-ai", "advanced-imessage", "package.json")
	if _, err := os.Stat(packagePath); err != nil {
		t.Skip("JS SDK is not installed; run `npm --prefix testdata/interop-js install` or `make test-interop`")
	}

	h := newInteropHarness(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	runGoInteropClient(t, ctx, h.address)
	goRecords := h.records()

	h.reset()
	runNodeInteropClient(t, ctx, h.address)
	jsRecords := h.records()

	compareInteropRecords(t, goRecords, jsRecords)
}

func runGoInteropClient(t *testing.T, ctx context.Context, address string) {
	t.Helper()

	client, err := New(address, StaticToken(interopToken), WithoutTLS())
	if err != nil {
		t.Fatalf("new Go client: %v", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			t.Fatalf("close Go client: %v", err)
		}
	}()

	chat := ChatGUID(interopChatGUID)
	replyTo := &ReplyTarget{GUID: interopReplyGUID, PartIndex: ptrInt(2)}
	formatting := []TextFormatInput{
		{Kind: FormatBold, Start: 0, Length: 5},
		{Kind: FormatEffect, Start: 6, Length: 5, Effect: TextEffectBloom},
	}

	msg, err := client.Messages().SendText(ctx, chat, "hello from interop", &SendTextOptions{
		ReplyTo:             replyTo,
		Subject:             "interop subject",
		Effect:              EffectConfetti,
		EnableDataDetection: true,
		EnableLinkPreview:   true,
		Formatting:          formatting,
		ClientMessageID:     "cmid-send-text",
	})
	if err != nil {
		t.Fatalf("SendText: %v", err)
	}
	assertGoFullMessage(t, msg)

	msg, err = client.Messages().SendAttachment(ctx, chat, interopAttachmentGUID, &SendAttachmentOptions{
		ReplyTo:         replyTo,
		Effect:          EffectGentle,
		IsAudioMessage:  true,
		ClientMessageID: "cmid-send-attachment",
	})
	if err != nil {
		t.Fatalf("SendAttachment: %v", err)
	}
	assertGoFullMessage(t, msg)

	msg, err = client.Messages().SendMultipart(ctx, chat, []MessagePart{
		{Text: "hello ", BubbleIndex: ptrInt(0), Formatting: formatting},
		{Text: "@Alice", MentionedAddress: "alice@example.com", BubbleIndex: ptrInt(1)},
		{AttachmentGUID: interopAttachmentGUID, AttachmentName: "photo.jpg", BubbleIndex: ptrInt(2)},
	}, &SendMultipartOptions{
		ReplyTo:             replyTo,
		Subject:             "multipart subject",
		Effect:              EffectConfetti,
		EnableDataDetection: true,
		ClientMessageID:     "cmid-send-multipart",
	})
	if err != nil {
		t.Fatalf("SendMultipart: %v", err)
	}
	assertGoFullMessage(t, msg)

	appStoreID := int64(1_234_567_890_123)
	msg, err = client.Messages().SendCustomizedMiniApp(ctx, chat, CustomizedMiniAppMessage{
		AppName:           "Interop App",
		ExtensionBundleID: "com.example.MessagesExtension",
		TeamID:            "TEAMID1234",
		URL:               "https://example.com/interop",
		AppStoreID:        &appStoreID,
		Layout: MiniAppLayout{
			Caption:            "Caption",
			Subcaption:         "Subcaption",
			TrailingCaption:    "Trailing Caption",
			TrailingSubcaption: "Trailing Subcaption",
			ImageTitle:         "Image Title",
			ImageSubtitle:      "Image Subtitle",
			Summary:            "Summary",
			Image:              []byte{1, 2, 3, 4},
		},
	}, &IdempotencyOptions{ClientMessageID: "cmid-mini-app"})
	if err != nil {
		t.Fatalf("SendCustomizedMiniApp: %v", err)
	}
	assertGoFullMessage(t, msg)

	msg, err = client.Messages().Edit(ctx, chat, interopMessageGUID, "edited text", &EditOptions{
		BackwardCompatText: "edited fallback",
		PartIndex:          ptrInt(2),
		ClientMessageID:    "cmid-edit",
	})
	if err != nil {
		t.Fatalf("Edit: %v", err)
	}
	assertGoFullMessage(t, msg)

	if err := client.Messages().Unsend(ctx, chat, interopMessageGUID, &PartOptions{
		PartIndex:       ptrInt(2),
		ClientMessageID: "cmid-unsend",
	}); err != nil {
		t.Fatalf("Unsend: %v", err)
	}

	reactions := []Reaction{
		{Kind: ReactionLove},
		{Kind: ReactionLike},
		{Kind: ReactionDislike},
		{Kind: ReactionLaugh},
		{Kind: ReactionEmphasize},
		{Kind: ReactionQuestion},
		{Kind: ReactionEmoji, Emoji: ":)"},
	}
	for i, reaction := range reactions {
		opts := &PartOptions{
			PartIndex:       ptrInt(2),
			ClientMessageID: fmt.Sprintf("cmid-reaction-%s", reaction.Kind),
		}
		// Alternate add/remove to exercise both is_set values on the wire.
		if i%2 == 0 {
			msg, err = client.Messages().AddReaction(ctx, chat, interopMessageGUID, reaction, opts)
		} else {
			msg, err = client.Messages().RemoveReaction(ctx, chat, interopMessageGUID, reaction, opts)
		}
		if err != nil {
			t.Fatalf("Add/RemoveReaction(%s): %v", reaction.Kind, err)
		}
		assertGoFullMessage(t, msg)
	}

	msg, err = client.Messages().PlaceSticker(ctx, chat, interopMessageGUID, StickerInput{
		GUID: interopStickerGUID,
		Placement: StickerPlacement{
			X:        0.4,
			Y:        0.6,
			Scale:    ptrFloat(1.25),
			Rotation: ptrFloat(0.5),
			Width:    ptrFloat(0.3),
		},
	}, &PartOptions{
		PartIndex:       ptrInt(2),
		ClientMessageID: "cmid-sticker",
	})
	if err != nil {
		t.Fatalf("PlaceSticker: %v", err)
	}
	assertGoFullMessage(t, msg)

	if err := client.Messages().NotifySilenced(ctx, chat, interopMessageGUID, &IdempotencyOptions{ClientMessageID: "cmid-notify"}); err != nil {
		t.Fatalf("NotifySilenced: %v", err)
	}

	msg, err = client.Messages().Get(ctx, interopMessageGUID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	assertGoFullMessage(t, msg)

	page, err := client.Messages().ListRecent(ctx, &MessageListFilter{
		Before:    ptrTime(interopBaseTime.Add(30 * time.Minute)),
		After:     ptrTime(interopBaseTime.Add(-30 * time.Minute)),
		IsFromMe:  ptrBool(true),
		IsRead:    ptrBool(false),
		PageSize:  25,
		PageToken: "recent-page-token",
	})
	if err != nil {
		t.Fatalf("ListRecent: %v", err)
	}
	assertGoPage(t, page)

	page, err = client.Messages().ListInChat(ctx, chat, &MessageListFilter{
		Before:    ptrTime(interopBaseTime.Add(45 * time.Minute)),
		After:     ptrTime(interopBaseTime.Add(-15 * time.Minute)),
		IsFromMe:  ptrBool(false),
		IsRead:    ptrBool(true),
		PageSize:  10,
		PageToken: "chat-page-token",
	})
	if err != nil {
		t.Fatalf("ListInChat: %v", err)
	}
	assertGoPage(t, page)

	media, err := client.Messages().GetEmbeddedMedia(ctx, chat, interopMessageGUID)
	if err != nil {
		t.Fatalf("GetEmbeddedMedia: %v", err)
	}
	if !reflect.DeepEqual(media.Data, []byte{9, 8, 7, 6}) || media.MIMEType != "image/png" {
		t.Fatalf("EmbeddedMedia = %+v", media)
	}

	sub := client.Messages().Subscribe(ctx, &SubscribeOptions{Chat: chat})
	messageEvents := collectGoEvents(t, sub.Events())
	if err := sub.Err(); err != nil {
		t.Fatalf("message subscription: %v", err)
	}
	if err := sub.Close(); err != nil {
		t.Fatalf("close message subscription: %v", err)
	}
	assertGoEventNames(t, messageEvents, []string{
		"message.received",
		"message.edited",
		"message.read",
		"message.unsent",
		"message.reactionAdded",
		"message.reactionRemoved",
		"message.stickerPlaced",
	})
	if received, ok := messageEvents[0].(MessageReceived); !ok {
		t.Fatalf("messageEvents[0] = %T, want MessageReceived", messageEvents[0])
	} else {
		assertGoFullMessage(t, received.Message)
	}

	catchUp := client.Events().CatchUp(ctx, 42)
	catchUpEvents := collectGoEvents(t, catchUp.Events())
	if err := catchUp.Err(); err != nil {
		t.Fatalf("catch-up stream: %v", err)
	}
	if err := catchUp.Close(); err != nil {
		t.Fatalf("close catch-up stream: %v", err)
	}
	assertGoEventNames(t, catchUpEvents, []string{
		"message.received",
		"message.edited",
		"message.read",
		"message.unsent",
		"message.reactionAdded",
		"message.reactionRemoved",
		"message.stickerPlaced",
		"chat.backgroundChanged",
		"chat.backgroundRemoved",
		"chat.markedRead",
		"chat.archived",
		"chat.unarchived",
		"group.changed.displayNameChanged",
		"group.changed.participantAdded",
		"group.changed.participantRemoved",
		"group.changed.participantLeft",
		"group.changed.iconChanged",
		"group.changed.iconRemoved",
		"poll.changed.created",
		"poll.changed.optionAdded",
		"poll.changed.voted",
		"poll.changed.unvoted",
		"catchup.complete",
	})
	if complete, ok := catchUpEvents[len(catchUpEvents)-1].(CatchupComplete); !ok || complete.HeadSequence != 3000 {
		t.Fatalf("last catch-up event = %#v, want CatchupComplete{HeadSequence:3000}", catchUpEvents[len(catchUpEvents)-1])
	}
}

func runNodeInteropClient(t *testing.T, ctx context.Context, address string) {
	t.Helper()

	cmd := exec.CommandContext(ctx, "node", filepath.Join("testdata", "interop-js", "client.mjs"), address)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("node interop client failed: %v\n%s", err, out)
	}
	t.Logf("node interop client: %s", out)
}

type interopRecord struct {
	Method  string
	Payload any
}

type interopServer struct {
	imessagev1connect.UnimplementedMessageServiceHandler
	imessagev1connect.UnimplementedEventServiceHandler

	mu      sync.Mutex
	records []interopRecord
}

func (s *interopServer) SendTextMessage(_ context.Context, req *connect.Request[imessagev1.SendTextMessageRequest]) (*connect.Response[imessagev1.MessageResponse], error) {
	s.record("SendTextMessage", req.Msg)
	return connect.NewResponse(interopMessageResponse()), nil
}

func (s *interopServer) SendAttachmentMessage(_ context.Context, req *connect.Request[imessagev1.SendAttachmentMessageRequest]) (*connect.Response[imessagev1.MessageResponse], error) {
	s.record("SendAttachmentMessage", req.Msg)
	return connect.NewResponse(interopMessageResponse()), nil
}

func (s *interopServer) SendMultipartMessage(_ context.Context, req *connect.Request[imessagev1.SendMultipartMessageRequest]) (*connect.Response[imessagev1.MessageResponse], error) {
	s.record("SendMultipartMessage", req.Msg)
	return connect.NewResponse(interopMessageResponse()), nil
}

func (s *interopServer) SendCustomizedMiniAppMessage(_ context.Context, req *connect.Request[imessagev1.SendCustomizedMiniAppMessageRequest]) (*connect.Response[imessagev1.MessageResponse], error) {
	s.record("SendCustomizedMiniAppMessage", req.Msg)
	return connect.NewResponse(interopMessageResponse()), nil
}

func (s *interopServer) EditMessage(_ context.Context, req *connect.Request[imessagev1.EditMessageRequest]) (*connect.Response[imessagev1.MessageResponse], error) {
	s.record("EditMessage", req.Msg)
	return connect.NewResponse(interopMessageResponse()), nil
}

func (s *interopServer) UnsendMessage(_ context.Context, req *connect.Request[imessagev1.UnsendMessageRequest]) (*connect.Response[emptypb.Empty], error) {
	s.record("UnsendMessage", req.Msg)
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (s *interopServer) SetReaction(_ context.Context, req *connect.Request[imessagev1.SetReactionRequest]) (*connect.Response[imessagev1.MessageResponse], error) {
	s.record("SetReaction", req.Msg)
	return connect.NewResponse(interopMessageResponse()), nil
}

func (s *interopServer) PlaceSticker(_ context.Context, req *connect.Request[imessagev1.PlaceStickerRequest]) (*connect.Response[imessagev1.MessageResponse], error) {
	s.record("PlaceSticker", req.Msg)
	return connect.NewResponse(interopMessageResponse()), nil
}

func (s *interopServer) NotifySilencedMessage(_ context.Context, req *connect.Request[imessagev1.NotifySilencedMessageRequest]) (*connect.Response[emptypb.Empty], error) {
	s.record("NotifySilencedMessage", req.Msg)
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (s *interopServer) GetMessage(_ context.Context, req *connect.Request[imessagev1.GetMessageRequest]) (*connect.Response[imessagev1.GetMessageResponse], error) {
	s.record("GetMessage", req.Msg)
	resp := &imessagev1.GetMessageResponse{}
	resp.SetMessage(interopMessage())
	return connect.NewResponse(resp), nil
}

func (s *interopServer) ListRecentMessages(_ context.Context, req *connect.Request[imessagev1.ListRecentMessagesRequest]) (*connect.Response[imessagev1.ListRecentMessagesResponse], error) {
	s.record("ListRecentMessages", req.Msg)
	resp := &imessagev1.ListRecentMessagesResponse{}
	resp.SetMessages([]*imessagev1.Message{interopMessage()})
	resp.SetNextPageToken("next-page-token")
	return connect.NewResponse(resp), nil
}

func (s *interopServer) ListChatMessages(_ context.Context, req *connect.Request[imessagev1.ListChatMessagesRequest]) (*connect.Response[imessagev1.ListChatMessagesResponse], error) {
	s.record("ListChatMessages", req.Msg)
	resp := &imessagev1.ListChatMessagesResponse{}
	resp.SetMessages([]*imessagev1.Message{interopMessage()})
	resp.SetNextPageToken("next-page-token")
	return connect.NewResponse(resp), nil
}

func (s *interopServer) GetEmbeddedMedia(_ context.Context, req *connect.Request[imessagev1.GetEmbeddedMediaRequest]) (*connect.Response[imessagev1.GetEmbeddedMediaResponse], error) {
	s.record("GetEmbeddedMedia", req.Msg)
	media := &imessagev1.EmbeddedMedia{}
	media.SetData([]byte{9, 8, 7, 6})
	media.SetMimeType("image/png")
	resp := &imessagev1.GetEmbeddedMediaResponse{}
	resp.SetMedia(media)
	return connect.NewResponse(resp), nil
}

func (s *interopServer) SubscribeMessageEvents(_ context.Context, req *connect.Request[imessagev1.SubscribeMessageEventsRequest], stream *connect.ServerStream[imessagev1.SubscribeMessageEventsResponse]) error {
	s.record("SubscribeMessageEvents", req.Msg)
	for i, ev := range interopMessageChangeEvents() {
		resp := &imessagev1.SubscribeMessageEventsResponse{}
		resp.SetSequence(uint64(1000 + i))
		resp.SetMessageChanged(ev)
		if err := stream.Send(resp); err != nil {
			return err
		}
	}
	return nil
}

func (s *interopServer) CatchUpEvents(_ context.Context, req *connect.Request[imessagev1.CatchUpEventsRequest], stream *connect.ServerStream[imessagev1.CatchUpEventsResponse]) error {
	s.record("CatchUpEvents", req.Msg)
	sequence := uint64(2000)
	for _, ev := range interopMessageChangeEvents() {
		resp := &imessagev1.CatchUpEventsResponse{}
		resp.SetSequence(sequence)
		resp.SetMessageChanged(ev)
		if err := stream.Send(resp); err != nil {
			return err
		}
		sequence++
	}
	for _, ev := range interopChatChangeEvents() {
		resp := &imessagev1.CatchUpEventsResponse{}
		resp.SetSequence(sequence)
		resp.SetChatChanged(ev)
		if err := stream.Send(resp); err != nil {
			return err
		}
		sequence++
	}
	for _, ev := range interopGroupChangeEvents() {
		resp := &imessagev1.CatchUpEventsResponse{}
		resp.SetSequence(sequence)
		resp.SetGroupChanged(ev)
		if err := stream.Send(resp); err != nil {
			return err
		}
		sequence++
	}
	for _, ev := range interopPollChangeEvents() {
		resp := &imessagev1.CatchUpEventsResponse{}
		resp.SetSequence(sequence)
		resp.SetPollChanged(ev)
		if err := stream.Send(resp); err != nil {
			return err
		}
		sequence++
	}

	complete := &imessagev1.CatchUpEventsComplete{}
	complete.SetHeadSequence(3000)
	resp := &imessagev1.CatchUpEventsResponse{}
	resp.SetSequence(sequence)
	resp.SetComplete(complete)
	return stream.Send(resp)
}

func (s *interopServer) record(method string, msg proto.Message) {
	payload := protoJSON(msg)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records = append(s.records, interopRecord{Method: method, Payload: payload})
}

func (s *interopServer) snapshot() []interopRecord {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]interopRecord, len(s.records))
	copy(out, s.records)
	return out
}

func (s *interopServer) reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records = nil
}

type interopHarness struct {
	address string
	server  *http.Server
	ln      net.Listener
	svc     *interopServer
}

func newInteropHarness(t *testing.T) *interopHarness {
	t.Helper()

	svc := &interopServer{}
	mux := http.NewServeMux()
	messagePath, messageHandler := imessagev1connect.NewMessageServiceHandler(svc)
	eventPath, eventHandler := imessagev1connect.NewEventServiceHandler(svc)
	mux.Handle(messagePath, messageHandler)
	mux.Handle(eventPath, eventHandler)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	server := &http.Server{Handler: h2c.NewHandler(mux, &http2.Server{})}
	h := &interopHarness{address: ln.Addr().String(), server: server, ln: ln, svc: svc}
	t.Cleanup(h.close)

	go func() {
		err := server.Serve(ln)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			t.Errorf("interop server: %v", err)
		}
	}()
	return h
}

func (h *interopHarness) close() {
	_ = h.server.Close()
	_ = h.ln.Close()
}

func (h *interopHarness) records() []interopRecord { return h.svc.snapshot() }
func (h *interopHarness) reset()                   { h.svc.reset() }

func protoJSON(msg proto.Message) any {
	b, err := protojson.MarshalOptions{UseProtoNames: true}.Marshal(msg)
	if err != nil {
		panic(err)
	}
	var payload any
	if err := json.Unmarshal(b, &payload); err != nil {
		panic(err)
	}
	return payload
}

func compareInteropRecords(t *testing.T, goRecords, jsRecords []interopRecord) {
	t.Helper()
	if len(goRecords) != len(jsRecords) {
		t.Fatalf("record count mismatch: Go=%d JS=%d\nGo=%s\nJS=%s", len(goRecords), len(jsRecords), prettyJSON(goRecords), prettyJSON(jsRecords))
	}
	for i := range goRecords {
		if goRecords[i].Method != jsRecords[i].Method {
			t.Fatalf("record %d method mismatch: Go=%s JS=%s", i, goRecords[i].Method, jsRecords[i].Method)
		}
		if !reflect.DeepEqual(goRecords[i].Payload, jsRecords[i].Payload) {
			t.Fatalf("%s payload mismatch\nGo=%s\nJS=%s", goRecords[i].Method, prettyJSON(goRecords[i].Payload), prettyJSON(jsRecords[i].Payload))
		}
	}
}

func prettyJSON(v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("%#v", v)
	}
	return string(b)
}

func interopMessageResponse() *imessagev1.MessageResponse {
	resp := &imessagev1.MessageResponse{}
	resp.SetMessage(interopMessage())
	return resp
}

func interopMessage() *imessagev1.Message {
	msg := &imessagev1.Message{}
	msg.SetGuid("fixture-message-guid")
	msg.SetContent(interopMessageContent("fixture body"))
	msg.SetSubject("fixture subject")
	msg.SetDateCreated(interopTS(0))
	msg.SetDateRead(interopTS(time.Minute))
	msg.SetDateDelivered(interopTS(2 * time.Minute))
	msg.SetDateEdited(interopTS(3 * time.Minute))
	msg.SetDateRetracted(interopTS(4 * time.Minute))
	msg.SetDatePlayed(interopTS(5 * time.Minute))
	msg.SetDateExpressiveSendPlayed(interopTS(6 * time.Minute))
	msg.SetSender(interopHandle("+15550001111"))
	msg.SetIsFromMe(true)
	msg.SetIsSent(true)
	msg.SetIsDelivered(true)
	msg.SetIsDeliveredQuietly(true)
	msg.SetDidNotifyRecipient(true)
	msg.SetSendErrorCode(12)
	msg.SetIsAudioMessage(true)
	msg.SetIsAutoReply(true)
	msg.SetIsSystemMessage(true)
	msg.SetIsForward(true)
	msg.SetIsDelayed(true)
	msg.SetIsSpam(true)
	msg.SetDataDetectorResultsPresent(true)
	msg.SetIsArchived(true)
	msg.SetIsServiceMessage(true)
	msg.SetIsCorrupt(true)
	msg.SetIsExpirable(true)
	msg.SetShareStatus(1)
	msg.SetShareDirection(2)
	msg.SetItemType(imessagev1.MessageItemType_MESSAGE_ITEM_TYPE_CHAT_ACTION)
	msg.SetGroupTitle("Fixture Group")
	msg.SetChatActionType(44)
	msg.SetThreadOriginatorGuid("thread-originator-guid")
	msg.SetThreadOriginatorPart("thread-originator-part")
	msg.SetReplyToGuid("reply-to-guid")
	msg.SetReplyTargetGuid("reply-target-guid")
	msg.SetDestinationCallerId("destination-caller-id")
	msg.SetPartCount(3)
	msg.SetCachedRoomNames("room-a,room-b")
	msg.SetReactionTargetGuid("reaction-target-guid")
	msg.SetReactionTargetPartIndex(2)
	msg.SetReaction(interopReaction(imessagev1.MessageReactionKind_MESSAGE_REACTION_KIND_EMOJI, ":)"))
	msg.SetReactionSelected(true)
	msg.SetAppliedReactions([]*imessagev1.MessageAppliedReaction{interopAppliedReaction()})
	msg.SetPlacedStickers([]*imessagev1.MessagePlacedSticker{interopPlacedSticker()})
	msg.SetChatGuids([]string{interopChatGUID, interopGroupChatGUID})
	return msg
}

func interopMessageContent(text string) *imessagev1.MessageContent {
	content := &imessagev1.MessageContent{}
	content.SetText(text)
	content.SetAttachments([]*imessagev1.AttachmentInfo{interopAttachmentInfo("fixture-attachment-guid")})

	bold := &imessagev1.TextFormat{}
	bold.SetType("bold")
	bold.SetStart(0)
	bold.SetLength(7)

	effect := &imessagev1.TextFormat{}
	effect.SetType("effect")
	effect.SetStart(8)
	effect.SetLength(4)
	effect.SetEffectName(string(TextEffectBloom))
	content.SetFormatting([]*imessagev1.TextFormat{bold, effect})

	mention := &imessagev1.MessageMention{}
	mention.SetAddress("alice@example.com")
	mention.SetStart(13)
	mention.SetLength(5)
	content.SetMentions([]*imessagev1.MessageMention{mention})
	content.SetBalloonBundleId("com.example.balloon")
	content.SetExpressiveSendStyleId(string(EffectConfetti))
	return content
}

func interopAttachmentInfo(guid string) *imessagev1.AttachmentInfo {
	info := &imessagev1.AttachmentInfo{}
	info.SetGuid(guid)
	info.SetOriginalGuid("original-attachment-guid")
	info.SetFileName("photo.jpg")
	info.SetMimeType("image/jpeg")
	info.SetUti("public.jpeg")
	info.SetTotalBytes(123456)
	info.SetIsOutgoing(true)
	info.SetTransferState(imessagev1.TransferState_TRANSFER_STATE_FINISHED)
	info.SetIsHidden(true)
	info.SetIsSticker(true)
	info.SetCompanionKind(imessagev1.CompanionKind_COMPANION_KIND_LIVE_PHOTO_VIDEO)
	return info
}

func interopAppliedReaction() *imessagev1.MessageAppliedReaction {
	applied := &imessagev1.MessageAppliedReaction{}
	applied.SetMessageGuid("fixture-message-guid")
	applied.SetTargetPartIndex(1)
	applied.SetReaction(interopReaction(imessagev1.MessageReactionKind_MESSAGE_REACTION_KIND_LOVE, ""))
	applied.SetSender(interopHandle("+15550002222"))
	applied.SetIsFromMe(false)
	applied.SetDateCreated(interopTS(7 * time.Minute))
	return applied
}

func interopPlacedSticker() *imessagev1.MessagePlacedSticker {
	placed := &imessagev1.MessagePlacedSticker{}
	placed.SetMessageGuid("fixture-message-guid")
	placed.SetTargetPartIndex(2)
	placed.SetSender(interopHandle("+15550003333"))
	placed.SetIsFromMe(true)
	placed.SetDateCreated(interopTS(8 * time.Minute))
	placed.SetSticker(interopAttachmentInfo("fixture-sticker-guid"))
	placed.SetPlacement(interopStickerPlacement())
	return placed
}

func interopStickerPlacement() *imessagev1.StickerPlacement {
	placement := &imessagev1.StickerPlacement{}
	placement.SetX(0.4)
	placement.SetY(0.6)
	placement.SetScale(1.25)
	placement.SetRotation(0.5)
	placement.SetWidth(0.3)
	return placement
}

func interopReaction(kind imessagev1.MessageReactionKind, emoji string) *imessagev1.MessageReaction {
	reaction := &imessagev1.MessageReaction{}
	reaction.SetKind(kind)
	if emoji != "" {
		reaction.SetEmoji(emoji)
	}
	return reaction
}

func interopHandle(address string) *imessagev1.SingleServiceAddressInfo {
	handle := &imessagev1.SingleServiceAddressInfo{}
	handle.SetAddress(address)
	handle.SetService(imessagev1.ChatServiceType_CHAT_SERVICE_TYPE_IMESSAGE)
	handle.SetCountry("us")
	return handle
}

func interopTS(offset time.Duration) *timestamppb.Timestamp {
	return timestamppb.New(interopBaseTime.Add(offset))
}

func interopMessageChangeEvents() []*imessagev1.MessageChangeEvent {
	received := interopMessageEvent()
	mr := &imessagev1.MessageReceived{}
	mr.SetMessage(interopMessage())
	received.SetMessageReceived(mr)

	edited := interopMessageEvent()
	me := &imessagev1.MessageEdited{}
	me.SetMessageGuid(interopMessageGUID)
	me.SetContent(interopMessageContent("edited fixture body"))
	me.SetEditedAt(interopTS(10 * time.Minute))
	edited.SetMessageEdited(me)

	read := interopMessageEvent()
	readPayload := &imessagev1.MessageRead{}
	readPayload.SetMessageGuid(interopMessageGUID)
	readPayload.SetReadAt(interopTS(11 * time.Minute))
	read.SetMessageRead(readPayload)

	unsent := interopMessageEvent()
	unsentPayload := &imessagev1.MessageUnsent{}
	unsentPayload.SetMessageGuid(interopMessageGUID)
	unsentPayload.SetRetractedAt(interopTS(12 * time.Minute))
	unsent.SetMessageUnsent(unsentPayload)

	reactionAdded := interopMessageEvent()
	addedPayload := &imessagev1.MessageReactionAdded{}
	addedPayload.SetMessageGuid(interopMessageGUID)
	addedPayload.SetTargetPartIndex(2)
	addedPayload.SetReaction(interopReaction(imessagev1.MessageReactionKind_MESSAGE_REACTION_KIND_LIKE, ""))
	reactionAdded.SetReactionAdded(addedPayload)

	reactionRemoved := interopMessageEvent()
	removedPayload := &imessagev1.MessageReactionRemoved{}
	removedPayload.SetMessageGuid(interopMessageGUID)
	removedPayload.SetTargetPartIndex(2)
	removedPayload.SetReaction(interopReaction(imessagev1.MessageReactionKind_MESSAGE_REACTION_KIND_QUESTION, ""))
	reactionRemoved.SetReactionRemoved(removedPayload)

	stickerPlaced := interopMessageEvent()
	stickerPayload := &imessagev1.StickerPlaced{}
	stickerPayload.SetMessageGuid(interopMessageGUID)
	stickerPayload.SetTargetPartIndex(2)
	stickerPayload.SetSticker(interopAttachmentInfo("event-sticker-guid"))
	stickerPayload.SetPlacement(interopStickerPlacement())
	stickerPlaced.SetStickerPlaced(stickerPayload)

	return []*imessagev1.MessageChangeEvent{
		received,
		edited,
		read,
		unsent,
		reactionAdded,
		reactionRemoved,
		stickerPlaced,
	}
}

func interopMessageEvent() *imessagev1.MessageChangeEvent {
	ev := &imessagev1.MessageChangeEvent{}
	ev.SetChatGuid(interopChatGUID)
	ev.SetOccurredAt(interopTS(9 * time.Minute))
	ev.SetActor(interopHandle("+15550004444"))
	ev.SetIsFromMe(true)
	return ev
}

func interopChatChangeEvents() []*imessagev1.ChatChangeEvent {
	backgroundChanged := interopChatEvent()
	backgroundChanged.SetBackgroundChanged(&imessagev1.ChatBackgroundChanged{})

	backgroundRemoved := interopChatEvent()
	backgroundRemoved.SetBackgroundRemoved(&imessagev1.ChatBackgroundRemoved{})

	markedRead := interopChatEvent()
	markedRead.SetMarkedRead(&imessagev1.ChatMarkedRead{})

	archived := interopChatEvent()
	archived.SetArchived(&imessagev1.ChatArchived{})

	unarchived := interopChatEvent()
	unarchived.SetUnarchived(&imessagev1.ChatUnarchived{})

	return []*imessagev1.ChatChangeEvent{
		backgroundChanged,
		backgroundRemoved,
		markedRead,
		archived,
		unarchived,
	}
}

func interopChatEvent() *imessagev1.ChatChangeEvent {
	ev := &imessagev1.ChatChangeEvent{}
	ev.SetChatGuid(interopChatGUID)
	ev.SetOccurredAt(interopTS(13 * time.Minute))
	ev.SetActor(interopHandle("+15550005555"))
	ev.SetIsFromMe(false)
	return ev
}

func interopGroupChangeEvents() []*imessagev1.GroupChangeEvent {
	displayName := interopGroupEvent()
	displayNamePayload := &imessagev1.GroupDisplayNameChanged{}
	displayNamePayload.SetDisplayName("Interop Group")
	displayName.SetDisplayNameChanged(displayNamePayload)

	added := interopGroupEvent()
	addedPayload := &imessagev1.GroupParticipantAdded{}
	addedPayload.SetParticipant(interopHandle("+15550006666"))
	added.SetParticipantAdded(addedPayload)

	removed := interopGroupEvent()
	removedPayload := &imessagev1.GroupParticipantRemoved{}
	removedPayload.SetParticipant(interopHandle("+15550007777"))
	removed.SetParticipantRemoved(removedPayload)

	left := interopGroupEvent()
	leftPayload := &imessagev1.GroupParticipantLeft{}
	leftPayload.SetParticipant(interopHandle("+15550008888"))
	left.SetParticipantLeft(leftPayload)

	iconChanged := interopGroupEvent()
	iconChanged.SetIconChanged(&imessagev1.GroupIconChanged{})

	iconRemoved := interopGroupEvent()
	iconRemoved.SetIconRemoved(&imessagev1.GroupIconRemoved{})

	return []*imessagev1.GroupChangeEvent{
		displayName,
		added,
		removed,
		left,
		iconChanged,
		iconRemoved,
	}
}

func interopGroupEvent() *imessagev1.GroupChangeEvent {
	ev := &imessagev1.GroupChangeEvent{}
	ev.SetChatGuid(interopGroupChatGUID)
	ev.SetOccurredAt(interopTS(14 * time.Minute))
	ev.SetActor(interopHandle("+15550009999"))
	ev.SetIsFromMe(true)
	return ev
}

func interopPollChangeEvents() []*imessagev1.PollChangeEvent {
	created := interopPollEvent()
	createdPayload := &imessagev1.PollCreated{}
	createdPayload.SetTitle("Lunch?")
	createdPayload.SetOptions(interopPollOptions())
	created.SetCreated(createdPayload)

	optionAdded := interopPollEvent()
	optionPayload := &imessagev1.PollOptionAdded{}
	optionPayload.SetTitle("Lunch?")
	optionPayload.SetOptions(interopPollOptions())
	optionAdded.SetOptionAdded(optionPayload)

	voted := interopPollEvent()
	votedPayload := &imessagev1.PollVoted{}
	votedPayload.SetOptionIdentifier("option-1")
	voted.SetVoted(votedPayload)

	unvoted := interopPollEvent()
	unvotedPayload := &imessagev1.PollUnvoted{}
	unvotedPayload.SetOptionIdentifier("option-2")
	unvoted.SetUnvoted(unvotedPayload)

	return []*imessagev1.PollChangeEvent{created, optionAdded, voted, unvoted}
}

func interopPollEvent() *imessagev1.PollChangeEvent {
	ev := &imessagev1.PollChangeEvent{}
	ev.SetChatGuid(interopChatGUID)
	ev.SetPollMessageGuid("poll-message-guid")
	ev.SetOccurredAt(interopTS(15 * time.Minute))
	ev.SetActor(interopHandle("+15550001010"))
	ev.SetIsFromMe(false)
	return ev
}

func interopPollOptions() []*imessagev1.PollOption {
	first := &imessagev1.PollOption{}
	first.SetOptionIdentifier("option-1")
	first.SetText("Sushi")
	first.SetCreatorHandle("alice@example.com")

	second := &imessagev1.PollOption{}
	second.SetOptionIdentifier("option-2")
	second.SetText("Pizza")
	second.SetCreatorHandle("bob@example.com")

	return []*imessagev1.PollOption{first, second}
}

func assertGoFullMessage(t *testing.T, msg Message) {
	t.Helper()

	if msg.GUID != "fixture-message-guid" {
		t.Fatalf("GUID = %q", msg.GUID)
	}
	if msg.Subject != "fixture subject" {
		t.Fatalf("Subject = %q", msg.Subject)
	}
	if !msg.DateCreated.Equal(interopBaseTime) {
		t.Fatalf("DateCreated = %v", msg.DateCreated)
	}
	assertTimePtr(t, msg.DateRead, interopBaseTime.Add(time.Minute), "DateRead")
	assertTimePtr(t, msg.DateDelivered, interopBaseTime.Add(2*time.Minute), "DateDelivered")
	assertTimePtr(t, msg.DateEdited, interopBaseTime.Add(3*time.Minute), "DateEdited")
	assertTimePtr(t, msg.DateRetracted, interopBaseTime.Add(4*time.Minute), "DateRetracted")
	assertTimePtr(t, msg.DatePlayed, interopBaseTime.Add(5*time.Minute), "DatePlayed")
	assertTimePtr(t, msg.DateExpressiveSendPlayed, interopBaseTime.Add(6*time.Minute), "DateExpressiveSendPlayed")
	if msg.Sender == nil || msg.Sender.Address != "+15550001111" || msg.Sender.Service != ServiceIMessage || msg.Sender.Country != "us" {
		t.Fatalf("Sender = %+v", msg.Sender)
	}
	if !msg.IsFromMe || !msg.IsSent || !msg.IsDelivered || !msg.IsDeliveredQuietly || !msg.DidNotifyRecipient {
		t.Fatalf("delivery flags not mapped: %+v", msg)
	}
	if msg.SendErrorCode != 12 || !msg.IsAudioMessage || !msg.IsAutoReply || !msg.IsSystemMessage || !msg.IsForward || !msg.IsDelayed || !msg.IsSpam {
		t.Fatalf("message flags not mapped: %+v", msg)
	}
	if !msg.DataDetectorResultsPresent || !msg.IsArchived || !msg.IsServiceMessage || !msg.IsCorrupt || !msg.IsExpirable {
		t.Fatalf("state flags not mapped: %+v", msg)
	}
	if msg.ShareStatus == nil || *msg.ShareStatus != 1 || msg.ShareDirection == nil || *msg.ShareDirection != 2 {
		t.Fatalf("share fields = status:%v direction:%v", msg.ShareStatus, msg.ShareDirection)
	}
	if msg.ItemType != ItemChatAction || msg.GroupTitle != "Fixture Group" || msg.ChatActionType == nil || *msg.ChatActionType != 44 {
		t.Fatalf("item fields not mapped: %+v", msg)
	}
	if msg.ThreadOriginatorGUID != "thread-originator-guid" || msg.ThreadOriginatorPart != "thread-originator-part" || msg.ReplyToGUID != "reply-to-guid" || msg.ReplyTargetGUID != "reply-target-guid" || msg.DestinationCallerID != "destination-caller-id" {
		t.Fatalf("thread/reply fields not mapped: %+v", msg)
	}
	if msg.PartCount == nil || *msg.PartCount != 3 || msg.CachedRoomNames != "room-a,room-b" {
		t.Fatalf("part fields not mapped: %+v", msg)
	}
	if msg.ReactionTargetGUID != "reaction-target-guid" || msg.ReactionTargetPartIndex == nil || *msg.ReactionTargetPartIndex != 2 {
		t.Fatalf("reaction target not mapped: %+v", msg)
	}
	if msg.Reaction == nil || msg.Reaction.Kind != ReactionEmoji || msg.Reaction.Emoji != ":)" {
		t.Fatalf("reaction = %+v", msg.Reaction)
	}
	if msg.ReactionSelected == nil || !*msg.ReactionSelected {
		t.Fatalf("ReactionSelected = %v", msg.ReactionSelected)
	}
	if !reflect.DeepEqual(msg.ChatGUIDs, []string{interopChatGUID, interopGroupChatGUID}) {
		t.Fatalf("ChatGUIDs = %v", msg.ChatGUIDs)
	}

	assertGoContent(t, msg.Content)
	if len(msg.AppliedReactions) != 1 {
		t.Fatalf("AppliedReactions length = %d", len(msg.AppliedReactions))
	}
	applied := msg.AppliedReactions[0]
	if applied.MessageGUID != "fixture-message-guid" || applied.TargetPartIndex == nil || *applied.TargetPartIndex != 1 || applied.Reaction.Kind != ReactionLove || applied.Sender == nil || applied.Sender.Address != "+15550002222" {
		t.Fatalf("AppliedReactions[0] = %+v", applied)
	}
	if !applied.DateCreated.Equal(interopBaseTime.Add(7 * time.Minute)) {
		t.Fatalf("AppliedReactions[0].DateCreated = %v", applied.DateCreated)
	}
	if len(msg.PlacedStickers) != 1 {
		t.Fatalf("PlacedStickers length = %d", len(msg.PlacedStickers))
	}
	sticker := msg.PlacedStickers[0]
	if sticker.MessageGUID != "fixture-message-guid" || sticker.TargetPartIndex == nil || *sticker.TargetPartIndex != 2 || sticker.Sticker == nil || sticker.Sticker.GUID != "fixture-sticker-guid" {
		t.Fatalf("PlacedStickers[0] = %+v", sticker)
	}
	assertGoPlacement(t, sticker.Placement)
}

func assertGoContent(t *testing.T, content MessageContent) {
	t.Helper()
	if content.Text != "fixture body" || content.BalloonBundleID != "com.example.balloon" || content.ExpressiveSendStyleID != string(EffectConfetti) {
		t.Fatalf("Content = %+v", content)
	}
	if len(content.Attachments) != 1 {
		t.Fatalf("Attachments length = %d", len(content.Attachments))
	}
	attachment := content.Attachments[0]
	if attachment.GUID != "fixture-attachment-guid" || attachment.OriginalGUID != "original-attachment-guid" || attachment.FileName != "photo.jpg" || attachment.MIMEType != "image/jpeg" || attachment.UTI != "public.jpeg" || attachment.TotalBytes != 123456 || !attachment.IsOutgoing || attachment.TransferState != TransferStateFinished || !attachment.IsHidden || !attachment.IsSticker || attachment.CompanionKind != CompanionLivePhotoVideo {
		t.Fatalf("Attachment = %+v", attachment)
	}
	if len(content.Formatting) != 2 || content.Formatting[0].Type != "bold" || content.Formatting[1].EffectName != string(TextEffectBloom) {
		t.Fatalf("Formatting = %+v", content.Formatting)
	}
	if len(content.Mentions) != 1 || content.Mentions[0].Address != "alice@example.com" {
		t.Fatalf("Mentions = %+v", content.Mentions)
	}
}

func assertGoPlacement(t *testing.T, placement *StickerPlacement) {
	t.Helper()
	if placement == nil || placement.X != 0.4 || placement.Y != 0.6 {
		t.Fatalf("Placement = %+v", placement)
	}
	if placement.Scale == nil || *placement.Scale != 1.25 || placement.Rotation == nil || *placement.Rotation != 0.5 || placement.Width == nil || *placement.Width != 0.3 {
		t.Fatalf("Placement optionals = %+v", placement)
	}
}

func assertGoPage(t *testing.T, page MessageListPage) {
	t.Helper()
	if page.NextPageToken != "next-page-token" || len(page.Messages) != 1 {
		t.Fatalf("MessageListPage = %+v", page)
	}
	assertGoFullMessage(t, page.Messages[0])
}

func assertTimePtr(t *testing.T, got *time.Time, want time.Time, name string) {
	t.Helper()
	if got == nil || !got.Equal(want) {
		t.Fatalf("%s = %v, want %v", name, got, want)
	}
}

func ptrTime(t time.Time) *time.Time { return &t }

func collectGoEvents[T any](t *testing.T, seq func(func(T) bool)) []T {
	t.Helper()
	var out []T
	for event := range seq {
		out = append(out, event)
	}
	return out
}

func assertGoEventNames[T any](t *testing.T, got []T, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("event count = %d, want %d: %v", len(got), len(want), got)
	}
	for i, event := range got {
		if name := goEventName(event); name != want[i] {
			t.Fatalf("event %d = %s (%T), want %s", i, name, event, want[i])
		}
	}
}

func goEventName(event any) string {
	switch e := event.(type) {
	case MessageReceived:
		return "message.received"
	case MessageEdited:
		return "message.edited"
	case MessageRead:
		return "message.read"
	case MessageUnsent:
		return "message.unsent"
	case MessageReactionAdded:
		return "message.reactionAdded"
	case MessageReactionRemoved:
		return "message.reactionRemoved"
	case MessageStickerPlaced:
		return "message.stickerPlaced"
	case ChatBackgroundChanged:
		return "chat.backgroundChanged"
	case ChatBackgroundRemoved:
		return "chat.backgroundRemoved"
	case ChatMarkedRead:
		return "chat.markedRead"
	case ChatArchived:
		return "chat.archived"
	case ChatUnarchived:
		return "chat.unarchived"
	case GroupChanged:
		switch e.Change.(type) {
		case GroupDisplayNameChanged:
			return "group.changed.displayNameChanged"
		case GroupParticipantAdded:
			return "group.changed.participantAdded"
		case GroupParticipantRemoved:
			return "group.changed.participantRemoved"
		case GroupParticipantLeft:
			return "group.changed.participantLeft"
		case GroupIconChanged:
			return "group.changed.iconChanged"
		case GroupIconRemoved:
			return "group.changed.iconRemoved"
		default:
			return fmt.Sprintf("group.changed.%T", e.Change)
		}
	case PollChanged:
		switch e.Delta.(type) {
		case PollCreated:
			return "poll.changed.created"
		case PollOptionAdded:
			return "poll.changed.optionAdded"
		case PollVoted:
			return "poll.changed.voted"
		case PollUnvoted:
			return "poll.changed.unvoted"
		default:
			return fmt.Sprintf("poll.changed.%T", e.Delta)
		}
	case CatchupComplete:
		return "catchup.complete"
	default:
		return fmt.Sprintf("%T", event)
	}
}
