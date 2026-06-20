package imessage

import (
	"testing"
	"time"

	imessagev1 "buf.build/gen/go/photon-hq/imessage/protocolbuffers/go/photon/imessage/v1"
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
}

func TestReactionKindRoundTrip(t *testing.T) {
	kinds := []MessageReactionKind{
		ReactionLove, ReactionLike, ReactionDislike, ReactionLaugh,
		ReactionEmphasize, ReactionQuestion, ReactionEmoji, ReactionSticker,
	}
	for _, k := range kinds {
		if got := reactionKindFromProto(reactionKindToProto(k)); got != k {
			t.Errorf("round trip %q -> %q", k, got)
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
