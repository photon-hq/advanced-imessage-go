package imessage

import (
	"encoding/base64"
	"testing"
	"time"

	imessagev1 "buf.build/gen/go/photon-hq/imessage/protocolbuffers/go/photon/imessage/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestMessageFromProtoMapsFields(t *testing.T) {
	created := time.Date(2026, 6, 20, 12, 0, 0, 0, time.UTC)
	read := created.Add(time.Minute)

	sender := &imessagev1.SingleServiceAddressInfo{}
	sender.SetAddress("+15550001111")
	sender.SetService(imessagev1.ChatServiceType_CHAT_SERVICE_TYPE_IMESSAGE)
	sender.SetCountry("us")

	content := &imessagev1.MessageContent{}
	content.SetText("hi there")

	reaction := &imessagev1.MessageReaction{}
	reaction.SetKind(imessagev1.MessageReactionKind_MESSAGE_REACTION_KIND_EMPHASIZE)

	pb := &imessagev1.Message{}
	pb.SetGuid("guid-1")
	pb.SetContent(content)
	pb.SetSubject("subj")
	pb.SetDateCreated(timestamppb.New(created))
	pb.SetDateRead(timestamppb.New(read))
	pb.SetSender(sender)
	pb.SetIsFromMe(true)
	pb.SetIsSent(true)
	pb.SetItemType(imessagev1.MessageItemType_MESSAGE_ITEM_TYPE_GROUP_NAME_CHANGE)
	pb.SetSendErrorCode(7)
	pb.SetPartCount(3)
	pb.SetReaction(reaction)
	pb.SetReactionSelected(true)
	pb.SetChatGuids([]string{"iMessage;-;+1", "iMessage;+;c"})

	got := messageFromProto(pb)

	if got.GUID != "guid-1" {
		t.Errorf("GUID = %q", got.GUID)
	}
	if got.Content.Text != "hi there" {
		t.Errorf("Content.Text = %q", got.Content.Text)
	}
	if got.Subject != "subj" {
		t.Errorf("Subject = %q", got.Subject)
	}
	if !got.DateCreated.Equal(created) {
		t.Errorf("DateCreated = %v, want %v", got.DateCreated, created)
	}
	if got.DateRead == nil || !got.DateRead.Equal(read) {
		t.Errorf("DateRead = %v, want %v", got.DateRead, read)
	}
	if got.DateDelivered != nil {
		t.Errorf("DateDelivered = %v, want nil", got.DateDelivered)
	}
	if got.Sender == nil || got.Sender.Address != "+15550001111" || got.Sender.Service != ServiceIMessage {
		t.Errorf("Sender = %+v", got.Sender)
	}
	if !got.IsFromMe || !got.IsSent {
		t.Errorf("flags IsFromMe=%v IsSent=%v", got.IsFromMe, got.IsSent)
	}
	if got.ItemType != ItemGroupNameChange {
		t.Errorf("ItemType = %q", got.ItemType)
	}
	if got.SendErrorCode != 7 {
		t.Errorf("SendErrorCode = %d", got.SendErrorCode)
	}
	if got.PartCount == nil || *got.PartCount != 3 {
		t.Errorf("PartCount = %v, want 3", got.PartCount)
	}
	if got.Reaction == nil || got.Reaction.Kind != ReactionEmphasize {
		t.Errorf("Reaction = %+v", got.Reaction)
	}
	if got.ReactionSelected == nil || !*got.ReactionSelected {
		t.Errorf("ReactionSelected = %v", got.ReactionSelected)
	}
	if len(got.ChatGUIDs) != 2 {
		t.Errorf("ChatGUIDs = %v", got.ChatGUIDs)
	}
	// An absent optional must stay nil, not a zero value.
	if got.ChatActionType != nil {
		t.Errorf("ChatActionType = %v, want nil", got.ChatActionType)
	}
	if got.Content.MiniApp != nil {
		t.Errorf("Content.MiniApp = %+v, want nil", got.Content.MiniApp)
	}
}

func TestCatchUpEventFromProtoMapsMiniApp(t *testing.T) {
	// Fixed v11.3.0 CatchUpEventsResponse frame containing
	// Message.content.mini_app. This catches field-number drift between the
	// server schema, generated client, and the public Go mapping.
	wire, err := base64.StdEncoding.DecodeString("CMsRUt8CChdpTWVzc2FnZTstOysxNTU1MTIzNDU2NxIGCIfH8dMGUrsCCrgCCg1taW5pLWFwcC1ndWlkEvMBOvABCgpQOFhUNjIzMlNMEiZjb2Rlcy5waG90b24uRXhhbXBsZS5NZXNzYWdlc0V4dGVuc2lvbhoLRXhhbXBsZSBBcHAiHmh0dHBzOi8vZXhhbXBsZS5jb20vY2FyZD9pZD00MiokOEQ4OTgwMzQtNDA3Qi00RkY1LTkxRTgtOURDMTg5MTFEQ0E5MNKF2MwEOAFCXwoHQ2FwdGlvbhIKU3ViY2FwdGlvbhoIVHJhaWxpbmciD1RyYWlsaW5nIGRldGFpbCoLSW1hZ2UgdGl0bGUyDkltYWdlIHN1YnRpdGxlOhBGYWxsYmFjayBzdW1tYXJ5UgYIh8fx0waiAQ4KDCsxNTU1MTIzNDU2N+IFF2lNZXNzYWdlOy07KzE1NTUxMjM0NTY3")
	if err != nil {
		t.Fatalf("decode fixture: %v", err)
	}

	var frame imessagev1.CatchUpEventsResponse
	if err := proto.Unmarshal(wire, &frame); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}
	event, ok, err := catchUpEventFromProto(&frame)
	if err != nil {
		t.Fatalf("map catch-up event: %v", err)
	}
	if !ok {
		t.Fatal("map catch-up event: ok = false")
	}
	received, ok := event.(MessageReceived)
	if !ok {
		t.Fatalf("event type = %T, want MessageReceived", event)
	}

	miniApp := received.Message.Content.MiniApp
	if miniApp == nil {
		t.Fatal("Message.Content.MiniApp = nil")
	}
	if miniApp.TeamID != "P8XT6232SL" || miniApp.ExtensionBundleID != "codes.photon.Example.MessagesExtension" {
		t.Fatalf("MiniApp identity = %+v", miniApp)
	}
	assertStringPtr(t, "MiniApp.AppName", miniApp.AppName, "Example App")
	assertStringPtr(t, "MiniApp.URL", miniApp.URL, "https://example.com/card?id=42")
	assertStringPtr(t, "MiniApp.SessionID", miniApp.SessionID, "8D898034-407B-4FF5-91E8-9DC18911DCA9")
	if miniApp.AppStoreID == nil || *miniApp.AppStoreID != 1234567890 {
		t.Fatalf("MiniApp.AppStoreID = %v", miniApp.AppStoreID)
	}
	if !miniApp.Live {
		t.Error("MiniApp.Live = false, want true")
	}
	if miniApp.Layout == nil {
		t.Fatal("MiniApp.Layout = nil")
	}
	assertStringPtr(t, "MiniApp.Layout.Caption", miniApp.Layout.Caption, "Caption")
	assertStringPtr(t, "MiniApp.Layout.Subcaption", miniApp.Layout.Subcaption, "Subcaption")
	assertStringPtr(t, "MiniApp.Layout.TrailingCaption", miniApp.Layout.TrailingCaption, "Trailing")
	assertStringPtr(t, "MiniApp.Layout.TrailingSubcaption", miniApp.Layout.TrailingSubcaption, "Trailing detail")
	assertStringPtr(t, "MiniApp.Layout.ImageTitle", miniApp.Layout.ImageTitle, "Image title")
	assertStringPtr(t, "MiniApp.Layout.ImageSubtitle", miniApp.Layout.ImageSubtitle, "Image subtitle")
	assertStringPtr(t, "MiniApp.Layout.Summary", miniApp.Layout.Summary, "Fallback summary")
}

func TestMessageContentFromProtoPreservesIdentityOnlyMiniApp(t *testing.T) {
	miniApp := &imessagev1.MiniAppContent{}
	miniApp.SetTeamId("P8XT6232SL")
	miniApp.SetExtensionBundleId("codes.photon.Example.MessagesExtension")
	content := &imessagev1.MessageContent{}
	content.SetMiniApp(miniApp)

	got := messageContentFromProto(content).MiniApp
	if got == nil {
		t.Fatal("MessageContent.MiniApp = nil")
	}
	if got.TeamID != "P8XT6232SL" || got.ExtensionBundleID != "codes.photon.Example.MessagesExtension" {
		t.Fatalf("MiniApp identity = %+v", got)
	}
	if got.AppName != nil || got.URL != nil || got.SessionID != nil || got.AppStoreID != nil || got.Layout != nil {
		t.Fatalf("MiniApp optional fields = %+v, want nil", got)
	}
}

func assertStringPtr(t *testing.T, name string, got *string, want string) {
	t.Helper()
	if got == nil || *got != want {
		t.Fatalf("%s = %v, want %q", name, got, want)
	}
}

func TestReactionKindRoundTrip(t *testing.T) {
	// Settable kinds map to proto (ok=true) and round-trip back unchanged.
	settable := []MessageReactionKind{
		ReactionLove, ReactionLike, ReactionDislike, ReactionLaugh,
		ReactionEmphasize, ReactionQuestion, ReactionEmoji,
	}
	for _, k := range settable {
		pb, ok := reactionKindToProto(k)
		if !ok {
			t.Errorf("reactionKindToProto(%q) ok = false, want true", k)
			continue
		}
		if got := reactionKindFromProto(pb); got != k {
			t.Errorf("round trip %q -> %q", k, got)
		}
	}
	// Non-settable kinds must be rejected (ok=false), never coerced to a default.
	for _, k := range []MessageReactionKind{ReactionSticker, ReactionUnknown, MessageReactionKind("bogus")} {
		if _, ok := reactionKindToProto(k); ok {
			t.Errorf("reactionKindToProto(%q) ok = true, want false", k)
		}
	}
}

func TestEnumFromProtoUnknownDefaults(t *testing.T) {
	if got := chatServiceFromProto(imessagev1.ChatServiceType_CHAT_SERVICE_TYPE_UNSPECIFIED); got != ServiceUnknown {
		t.Errorf("unspecified service = %q, want %q", got, ServiceUnknown)
	}
	if got := transferStateFromProto(imessagev1.TransferState(999)); got != TransferStateUnknown {
		t.Errorf("unknown transfer state = %q, want %q", got, TransferStateUnknown)
	}
	if got := locationTypeFromProto(imessagev1.FriendLocationType_FRIEND_LOCATION_TYPE_UNSPECIFIED); got != LocationUnknown {
		t.Errorf("unspecified location type = %q, want %q", got, LocationUnknown)
	}
}
