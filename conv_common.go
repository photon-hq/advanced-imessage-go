package imessage

import (
	"time"

	imessagev1 "buf.build/gen/go/photon-hq/imessage/protocolbuffers/go/photon/imessage/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// This file holds the shared, unexported proto<->domain conversion helpers used
// by the per-resource conv_*.go files. Generated protobuf types never escape
// these helpers into the public API.

func timeFromProto(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Time{}
	}
	return ts.AsTime()
}

func timePtrFromProto(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	t := ts.AsTime()
	return &t
}

func protoTimestamp(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}

func ptrInt(v int32) *int {
	i := int(v)
	return &i
}

func ptrBool(v bool) *bool { return &v }

func ptrFloat(v float64) *float64 { return &v }

func handleFromProto(pb *imessagev1.SingleServiceAddressInfo) Handle {
	if pb == nil {
		return Handle{}
	}
	return Handle{
		Address: pb.GetAddress(),
		Service: chatServiceFromProto(pb.GetService()),
		Country: pb.GetCountry(),
	}
}

func handlePtrFromProto(pb *imessagev1.SingleServiceAddressInfo) *Handle {
	if pb == nil {
		return nil
	}
	h := handleFromProto(pb)
	return &h
}

func addressInfoFromProto(pb *imessagev1.MultiServiceAddressInfo) AddressInfo {
	if pb == nil {
		return AddressInfo{}
	}
	services := make([]ChatServiceType, 0, len(pb.GetServices()))
	for _, s := range pb.GetServices() {
		services = append(services, chatServiceFromProto(s))
	}
	return AddressInfo{
		Address:  pb.GetAddress(),
		Services: services,
		Country:  pb.GetCountry(),
	}
}

func chatServiceFromProto(s imessagev1.ChatServiceType) ChatServiceType {
	switch s {
	case imessagev1.ChatServiceType_CHAT_SERVICE_TYPE_IMESSAGE:
		return ServiceIMessage
	case imessagev1.ChatServiceType_CHAT_SERVICE_TYPE_SMS:
		return ServiceSMS
	case imessagev1.ChatServiceType_CHAT_SERVICE_TYPE_RCS:
		return ServiceRCS
	default:
		return ServiceUnknown
	}
}

func chatServiceToProto(s ChatServiceType) imessagev1.ChatServiceType {
	switch s {
	case ServiceIMessage:
		return imessagev1.ChatServiceType_CHAT_SERVICE_TYPE_IMESSAGE
	case ServiceSMS:
		return imessagev1.ChatServiceType_CHAT_SERVICE_TYPE_SMS
	case ServiceRCS:
		return imessagev1.ChatServiceType_CHAT_SERVICE_TYPE_RCS
	default:
		return imessagev1.ChatServiceType_CHAT_SERVICE_TYPE_UNSPECIFIED
	}
}

func transferStateFromProto(s imessagev1.TransferState) TransferState {
	switch s {
	case imessagev1.TransferState_TRANSFER_STATE_PENDING:
		return TransferStatePending
	case imessagev1.TransferState_TRANSFER_STATE_TRANSFERRING:
		return TransferStateTransferring
	case imessagev1.TransferState_TRANSFER_STATE_FAILED:
		return TransferStateFailed
	case imessagev1.TransferState_TRANSFER_STATE_FINISHED:
		return TransferStateFinished
	case imessagev1.TransferState_TRANSFER_STATE_UNAVAILABLE:
		return TransferStateUnavailable
	default:
		return TransferStateUnknown
	}
}

func companionKindFromProto(k imessagev1.CompanionKind) CompanionKind {
	switch k {
	case imessagev1.CompanionKind_COMPANION_KIND_LIVE_PHOTO_VIDEO:
		return CompanionLivePhotoVideo
	default:
		return CompanionUnknown
	}
}

func companionKindToProto(k CompanionKind) imessagev1.CompanionKind {
	switch k {
	case CompanionLivePhotoVideo:
		return imessagev1.CompanionKind_COMPANION_KIND_LIVE_PHOTO_VIDEO
	default:
		return imessagev1.CompanionKind_COMPANION_KIND_UNSPECIFIED
	}
}

func messageItemTypeFromProto(t imessagev1.MessageItemType) MessageItemType {
	switch t {
	case imessagev1.MessageItemType_MESSAGE_ITEM_TYPE_NORMAL:
		return ItemNormal
	case imessagev1.MessageItemType_MESSAGE_ITEM_TYPE_PARTICIPANT_CHANGE:
		return ItemParticipantChange
	case imessagev1.MessageItemType_MESSAGE_ITEM_TYPE_GROUP_NAME_CHANGE:
		return ItemGroupNameChange
	case imessagev1.MessageItemType_MESSAGE_ITEM_TYPE_CHAT_ACTION:
		return ItemChatAction
	default:
		return ItemUnknown
	}
}

func reactionKindFromProto(k imessagev1.MessageReactionKind) MessageReactionKind {
	switch k {
	case imessagev1.MessageReactionKind_MESSAGE_REACTION_KIND_LOVE:
		return ReactionLove
	case imessagev1.MessageReactionKind_MESSAGE_REACTION_KIND_LIKE:
		return ReactionLike
	case imessagev1.MessageReactionKind_MESSAGE_REACTION_KIND_DISLIKE:
		return ReactionDislike
	case imessagev1.MessageReactionKind_MESSAGE_REACTION_KIND_LAUGH:
		return ReactionLaugh
	case imessagev1.MessageReactionKind_MESSAGE_REACTION_KIND_EMPHASIZE:
		return ReactionEmphasize
	case imessagev1.MessageReactionKind_MESSAGE_REACTION_KIND_QUESTION:
		return ReactionQuestion
	case imessagev1.MessageReactionKind_MESSAGE_REACTION_KIND_EMOJI:
		return ReactionEmoji
	case imessagev1.MessageReactionKind_MESSAGE_REACTION_KIND_STICKER:
		return ReactionSticker
	default:
		return ReactionUnknown
	}
}

func reactionKindToProto(k MessageReactionKind) imessagev1.MessageReactionKind {
	switch k {
	case ReactionLove:
		return imessagev1.MessageReactionKind_MESSAGE_REACTION_KIND_LOVE
	case ReactionLike:
		return imessagev1.MessageReactionKind_MESSAGE_REACTION_KIND_LIKE
	case ReactionDislike:
		return imessagev1.MessageReactionKind_MESSAGE_REACTION_KIND_DISLIKE
	case ReactionLaugh:
		return imessagev1.MessageReactionKind_MESSAGE_REACTION_KIND_LAUGH
	case ReactionEmphasize:
		return imessagev1.MessageReactionKind_MESSAGE_REACTION_KIND_EMPHASIZE
	case ReactionQuestion:
		return imessagev1.MessageReactionKind_MESSAGE_REACTION_KIND_QUESTION
	case ReactionEmoji:
		return imessagev1.MessageReactionKind_MESSAGE_REACTION_KIND_EMOJI
	case ReactionSticker:
		return imessagev1.MessageReactionKind_MESSAGE_REACTION_KIND_STICKER
	default:
		return imessagev1.MessageReactionKind_MESSAGE_REACTION_KIND_LOVE
	}
}

func locationTypeFromProto(t imessagev1.FriendLocationType) LocationType {
	switch t {
	case imessagev1.FriendLocationType_FRIEND_LOCATION_TYPE_LEGACY:
		return LocationLegacy
	case imessagev1.FriendLocationType_FRIEND_LOCATION_TYPE_LIVE:
		return LocationLive
	case imessagev1.FriendLocationType_FRIEND_LOCATION_TYPE_SHALLOW:
		return LocationShallow
	default:
		return LocationUnknown
	}
}

func reactionFromProto(pb *imessagev1.MessageReaction) MessageReaction {
	if pb == nil {
		return MessageReaction{}
	}
	return MessageReaction{
		Kind:  reactionKindFromProto(pb.GetKind()),
		Emoji: pb.GetEmoji(),
	}
}
