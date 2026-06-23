package imessage

import (
	"cmp"
	"slices"

	imessagev1 "buf.build/gen/go/photon-hq/imessage/protocolbuffers/go/photon/imessage/v1"
)

// --- proto -> domain ---

func messageFromProto(pb *imessagev1.Message) Message {
	if pb == nil {
		return Message{}
	}
	m := Message{
		GUID:                       pb.GetGuid(),
		Content:                    messageContentFromProto(pb.GetContent()),
		Subject:                    pb.GetSubject(),
		DateCreated:                timeFromProto(pb.GetDateCreated()),
		DateRead:                   timePtrFromProto(pb.GetDateRead()),
		DateDelivered:              timePtrFromProto(pb.GetDateDelivered()),
		DateEdited:                 timePtrFromProto(pb.GetDateEdited()),
		DateRetracted:              timePtrFromProto(pb.GetDateRetracted()),
		DatePlayed:                 timePtrFromProto(pb.GetDatePlayed()),
		DateExpressiveSendPlayed:   timePtrFromProto(pb.GetDateExpressiveSendPlayed()),
		Sender:                     handlePtrFromProto(pb.GetSender()),
		IsFromMe:                   pb.GetIsFromMe(),
		IsSent:                     pb.GetIsSent(),
		IsDelivered:                pb.GetIsDelivered(),
		IsDeliveredQuietly:         pb.GetIsDeliveredQuietly(),
		DidNotifyRecipient:         pb.GetDidNotifyRecipient(),
		SendErrorCode:              int(pb.GetSendErrorCode()),
		IsAudioMessage:             pb.GetIsAudioMessage(),
		IsAutoReply:                pb.GetIsAutoReply(),
		IsSystemMessage:            pb.GetIsSystemMessage(),
		IsForward:                  pb.GetIsForward(),
		IsDelayed:                  pb.GetIsDelayed(),
		IsSpam:                     pb.GetIsSpam(),
		DataDetectorResultsPresent: pb.GetDataDetectorResultsPresent(),
		IsArchived:                 pb.GetIsArchived(),
		IsServiceMessage:           pb.GetIsServiceMessage(),
		IsCorrupt:                  pb.GetIsCorrupt(),
		IsExpirable:                pb.GetIsExpirable(),
		ItemType:                   messageItemTypeFromProto(pb.GetItemType()),
		GroupTitle:                 pb.GetGroupTitle(),
		ThreadOriginatorGUID:       pb.GetThreadOriginatorGuid(),
		ThreadOriginatorPart:       pb.GetThreadOriginatorPart(),
		ReplyToGUID:                pb.GetReplyToGuid(),
		ReplyTargetGUID:            pb.GetReplyTargetGuid(),
		DestinationCallerID:        pb.GetDestinationCallerId(),
		CachedRoomNames:            pb.GetCachedRoomNames(),
		ReactionTargetGUID:         pb.GetReactionTargetGuid(),
		ChatGUIDs:                  pb.GetChatGuids(),
		AppliedReactions:           appliedReactionsFromProto(pb.GetAppliedReactions()),
		PlacedStickers:             placedStickersFromProto(pb.GetPlacedStickers()),
	}
	if pb.HasShareStatus() {
		m.ShareStatus = ptrInt(pb.GetShareStatus())
	}
	if pb.HasShareDirection() {
		m.ShareDirection = ptrInt(pb.GetShareDirection())
	}
	if pb.HasChatActionType() {
		m.ChatActionType = ptrInt(pb.GetChatActionType())
	}
	if pb.HasPartCount() {
		m.PartCount = ptrInt(pb.GetPartCount())
	}
	if pb.HasReactionTargetPartIndex() {
		m.ReactionTargetPartIndex = ptrInt(pb.GetReactionTargetPartIndex())
	}
	if pb.HasReaction() {
		r := reactionFromProto(pb.GetReaction())
		m.Reaction = &r
	}
	if pb.HasReactionSelected() {
		m.ReactionSelected = ptrBool(pb.GetReactionSelected())
	}
	return m
}

func messagePtrFromProto(pb *imessagev1.Message) *Message {
	if pb == nil {
		return nil
	}
	m := messageFromProto(pb)
	return &m
}

func messageContentFromProto(pb *imessagev1.MessageContent) MessageContent {
	if pb == nil {
		return MessageContent{}
	}
	c := MessageContent{
		Text:                  pb.GetText(),
		BalloonBundleID:       pb.GetBalloonBundleId(),
		ExpressiveSendStyleID: pb.GetExpressiveSendStyleId(),
	}
	for _, a := range pb.GetAttachments() {
		c.Attachments = append(c.Attachments, attachmentInfoFromProto(a))
	}
	for _, f := range pb.GetFormatting() {
		c.Formatting = append(c.Formatting, textFormatFromProto(f))
	}
	for _, mn := range pb.GetMentions() {
		c.Mentions = append(c.Mentions, mentionFromProto(mn))
	}
	return c
}

func textFormatFromProto(pb *imessagev1.TextFormat) TextFormat {
	return TextFormat{
		Type:       pb.GetType(),
		Start:      int(pb.GetStart()),
		Length:     int(pb.GetLength()),
		EffectName: pb.GetEffectName(),
	}
}

func mentionFromProto(pb *imessagev1.MessageMention) MessageMention {
	return MessageMention{
		Address: pb.GetAddress(),
		Start:   int(pb.GetStart()),
		Length:  int(pb.GetLength()),
	}
}

func appliedReactionsFromProto(in []*imessagev1.MessageAppliedReaction) []MessageAppliedReaction {
	if len(in) == 0 {
		return nil
	}
	out := make([]MessageAppliedReaction, 0, len(in))
	for _, pb := range in {
		ar := MessageAppliedReaction{
			MessageGUID: pb.GetMessageGuid(),
			Reaction:    reactionFromProto(pb.GetReaction()),
			Sender:      handlePtrFromProto(pb.GetSender()),
			IsFromMe:    pb.GetIsFromMe(),
			DateCreated: timeFromProto(pb.GetDateCreated()),
		}
		if pb.HasTargetPartIndex() {
			ar.TargetPartIndex = ptrInt(pb.GetTargetPartIndex())
		}
		out = append(out, ar)
	}
	// Deterministic oldest-first order (DateCreated, then MessageGUID), owned by
	// the client; see the contract on [Message.AppliedReactions].
	slices.SortFunc(out, func(a, b MessageAppliedReaction) int {
		if c := a.DateCreated.Compare(b.DateCreated); c != 0 {
			return c
		}
		return cmp.Compare(a.MessageGUID, b.MessageGUID)
	})
	return out
}

func placedStickersFromProto(in []*imessagev1.MessagePlacedSticker) []MessagePlacedSticker {
	if len(in) == 0 {
		return nil
	}
	out := make([]MessagePlacedSticker, 0, len(in))
	for _, pb := range in {
		ps := MessagePlacedSticker{
			MessageGUID: pb.GetMessageGuid(),
			Sender:      handlePtrFromProto(pb.GetSender()),
			IsFromMe:    pb.GetIsFromMe(),
			DateCreated: timeFromProto(pb.GetDateCreated()),
			Sticker:     attachmentInfoPtrFromProto(pb.GetSticker()),
			Placement:   stickerPlacementFromProto(pb.GetPlacement()),
		}
		if pb.HasTargetPartIndex() {
			ps.TargetPartIndex = ptrInt(pb.GetTargetPartIndex())
		}
		out = append(out, ps)
	}
	// Deterministic oldest-first order (DateCreated, then MessageGUID), owned by
	// the client; see the contract on [Message.PlacedStickers].
	slices.SortFunc(out, func(a, b MessagePlacedSticker) int {
		if c := a.DateCreated.Compare(b.DateCreated); c != 0 {
			return c
		}
		return cmp.Compare(a.MessageGUID, b.MessageGUID)
	})
	return out
}

func stickerPlacementFromProto(pb *imessagev1.StickerPlacement) *StickerPlacement {
	if pb == nil {
		return nil
	}
	sp := &StickerPlacement{X: pb.GetX(), Y: pb.GetY()}
	if pb.HasScale() {
		sp.Scale = ptrFloat(pb.GetScale())
	}
	if pb.HasRotation() {
		sp.Rotation = ptrFloat(pb.GetRotation())
	}
	if pb.HasWidth() {
		sp.Width = ptrFloat(pb.GetWidth())
	}
	return sp
}

// messageEventFromChange maps a MessageChangeEvent (used by both the subscribe
// and catch-up streams) into a domain [MessageEvent].
func messageEventFromChange(ev *imessagev1.MessageChangeEvent, sequence uint64) (MessageEvent, bool) {
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
	case imessagev1.MessageChangeEvent_MessageReceived_case:
		return MessageReceived{EventContext: ctx, Sequence: sequence, Message: messageFromProto(ev.GetMessageReceived().GetMessage())}, true
	case imessagev1.MessageChangeEvent_MessageEdited_case:
		e := ev.GetMessageEdited()
		return MessageEdited{EventContext: ctx, Sequence: sequence, MessageGUID: e.GetMessageGuid(), Content: messageContentFromProto(e.GetContent()), EditedAt: timeFromProto(e.GetEditedAt())}, true
	case imessagev1.MessageChangeEvent_MessageRead_case:
		e := ev.GetMessageRead()
		return MessageRead{EventContext: ctx, Sequence: sequence, MessageGUID: e.GetMessageGuid(), ReadAt: timeFromProto(e.GetReadAt())}, true
	case imessagev1.MessageChangeEvent_MessageUnsent_case:
		e := ev.GetMessageUnsent()
		return MessageUnsent{EventContext: ctx, Sequence: sequence, MessageGUID: e.GetMessageGuid(), RetractedAt: timeFromProto(e.GetRetractedAt())}, true
	case imessagev1.MessageChangeEvent_ReactionAdded_case:
		e := ev.GetReactionAdded()
		ra := MessageReactionAdded{EventContext: ctx, Sequence: sequence, MessageGUID: e.GetMessageGuid(), Reaction: reactionFromProto(e.GetReaction())}
		if e.HasTargetPartIndex() {
			ra.TargetPartIndex = ptrInt(e.GetTargetPartIndex())
		}
		return ra, true
	case imessagev1.MessageChangeEvent_ReactionRemoved_case:
		e := ev.GetReactionRemoved()
		rr := MessageReactionRemoved{EventContext: ctx, Sequence: sequence, MessageGUID: e.GetMessageGuid(), Reaction: reactionFromProto(e.GetReaction())}
		if e.HasTargetPartIndex() {
			rr.TargetPartIndex = ptrInt(e.GetTargetPartIndex())
		}
		return rr, true
	case imessagev1.MessageChangeEvent_StickerPlaced_case:
		e := ev.GetStickerPlaced()
		sp := MessageStickerPlaced{EventContext: ctx, Sequence: sequence, MessageGUID: e.GetMessageGuid(), Sticker: attachmentInfoPtrFromProto(e.GetSticker()), Placement: stickerPlacementFromProto(e.GetPlacement())}
		if e.HasTargetPartIndex() {
			sp.TargetPartIndex = ptrInt(e.GetTargetPartIndex())
		}
		return sp, true
	default:
		return nil, false
	}
}

func messageEventFromProto(pb *imessagev1.SubscribeMessageEventsResponse) (MessageEvent, bool, error) {
	if pb.WhichPayload() != imessagev1.SubscribeMessageEventsResponse_MessageChanged_case {
		return nil, false, nil
	}
	ev, ok := messageEventFromChange(pb.GetMessageChanged(), pb.GetSequence())
	return ev, ok, nil
}

// --- domain -> proto ---

func replyTargetToProto(rt *ReplyTarget) *imessagev1.ReplyTarget {
	if rt == nil {
		return nil
	}
	pb := &imessagev1.ReplyTarget{}
	pb.SetMessageGuid(rt.GUID)
	if rt.PartIndex != nil {
		pb.SetTargetPartIndex(int32(*rt.PartIndex))
	}
	return pb
}

func textFormatInputToProto(f TextFormatInput) *imessagev1.TextFormat {
	pb := &imessagev1.TextFormat{}
	pb.SetType(string(f.Kind))
	pb.SetStart(int32(f.Start))
	pb.SetLength(int32(f.Length))
	if f.Kind == FormatEffect && f.Effect != "" {
		pb.SetEffectName(string(f.Effect))
	}
	return pb
}

func textFormatInputsToProto(in []TextFormatInput) []*imessagev1.TextFormat {
	if len(in) == 0 {
		return nil
	}
	out := make([]*imessagev1.TextFormat, 0, len(in))
	for _, f := range in {
		out = append(out, textFormatInputToProto(f))
	}
	return out
}

func messagePartToProto(p MessagePart) *imessagev1.MessagePart {
	pb := &imessagev1.MessagePart{}
	if p.Text != "" {
		pb.SetText(p.Text)
	}
	if p.AttachmentGUID != "" {
		ref := &imessagev1.AttachmentRef{}
		ref.SetAttachmentGuid(p.AttachmentGUID)
		if p.AttachmentName != "" {
			ref.SetAttachmentName(p.AttachmentName)
		}
		pb.SetAttachment(ref)
	}
	if p.MentionedAddress != "" {
		pb.SetMentionedAddress(p.MentionedAddress)
	}
	if p.BubbleIndex != nil {
		pb.SetBubbleIndex(int32(*p.BubbleIndex))
	}
	if len(p.Formatting) > 0 {
		pb.SetFormatting(textFormatInputsToProto(p.Formatting))
	}
	return pb
}

func stickerPlacementToProto(p StickerPlacement) *imessagev1.StickerPlacement {
	pb := &imessagev1.StickerPlacement{}
	pb.SetX(p.X)
	pb.SetY(p.Y)
	if p.Scale != nil {
		pb.SetScale(*p.Scale)
	}
	if p.Rotation != nil {
		pb.SetRotation(*p.Rotation)
	}
	if p.Width != nil {
		pb.SetWidth(*p.Width)
	}
	return pb
}

func messageTarget(chat ChatGUID, messageGUID string, partIndex *int) *imessagev1.MessageTarget {
	t := &imessagev1.MessageTarget{}
	t.SetChatGuid(chat.String())
	t.SetMessageGuid(messageGUID)
	if partIndex != nil {
		t.SetTargetPartIndex(int32(*partIndex))
	}
	return t
}
