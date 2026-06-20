package imessage

import (
	imessagev1 "buf.build/gen/go/photon-hq/imessage/protocolbuffers/go/photon/imessage/v1"
)

func chatFromProto(pb *imessagev1.Chat) Chat {
	if pb == nil {
		return Chat{}
	}
	c := Chat{
		GUID:           ChatGUID(pb.GetGuid()),
		ChatIdentifier: pb.GetChatIdentifier(),
		GroupID:        pb.GetGroupId(),
		DisplayName:    pb.GetDisplayName(),
		IsGroup:        pb.GetIsGroup(),
		Service:        chatServiceFromProto(pb.GetService()),
		IsArchived:     pb.GetIsArchived(),
		IsFiltered:     pb.GetIsFiltered(),
		LastMessage:    messagePtrFromProto(pb.GetLastMessage()),
	}
	if pb.HasUnreadCount() {
		c.UnreadCount = ptrInt(pb.GetUnreadCount())
	}
	for _, p := range pb.GetParticipants() {
		c.Participants = append(c.Participants, handleFromProto(p))
	}
	return c
}

func chatEventFromChange(ev *imessagev1.ChatChangeEvent, sequence uint64) (ChatEvent, bool) {
	if ev == nil {
		return nil, false
	}
	ctx := EventContext{
		Actor:      handlePtrFromProto(ev.GetActor()),
		ChatGUID:   ev.GetChatGuid(),
		IsFromMe:   ev.GetIsFromMe(),
		OccurredAt: timeFromProto(ev.GetOccurredAt()),
	}
	switch ev.WhichChange() {
	case imessagev1.ChatChangeEvent_BackgroundChanged_case:
		return ChatBackgroundChanged{EventContext: ctx, Sequence: sequence}, true
	case imessagev1.ChatChangeEvent_BackgroundRemoved_case:
		return ChatBackgroundRemoved{EventContext: ctx, Sequence: sequence}, true
	case imessagev1.ChatChangeEvent_MarkedRead_case:
		return ChatMarkedRead{EventContext: ctx, Sequence: sequence}, true
	case imessagev1.ChatChangeEvent_Archived_case:
		return ChatArchived{EventContext: ctx, Sequence: sequence}, true
	case imessagev1.ChatChangeEvent_Unarchived_case:
		return ChatUnarchived{EventContext: ctx, Sequence: sequence}, true
	default:
		return nil, false
	}
}

func chatEventFromProto(pb *imessagev1.SubscribeChatEventsResponse) (ChatEvent, bool, error) {
	if pb.WhichPayload() != imessagev1.SubscribeChatEventsResponse_ChatChanged_case {
		return nil, false, nil
	}
	ev, ok := chatEventFromChange(pb.GetChatChanged(), pb.GetSequence())
	return ev, ok, nil
}
