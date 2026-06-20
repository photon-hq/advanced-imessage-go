package imessage

import (
	imessagev1 "buf.build/gen/go/photon-hq/imessage/protocolbuffers/go/photon/imessage/v1"
)

func groupChangeFromProto(ev *imessagev1.GroupChangeEvent) (GroupChange, bool) {
	switch ev.WhichChange() {
	case imessagev1.GroupChangeEvent_DisplayNameChanged_case:
		return GroupDisplayNameChanged{DisplayName: ev.GetDisplayNameChanged().GetDisplayName()}, true
	case imessagev1.GroupChangeEvent_ParticipantAdded_case:
		return GroupParticipantAdded{Participant: handleFromProto(ev.GetParticipantAdded().GetParticipant())}, true
	case imessagev1.GroupChangeEvent_ParticipantRemoved_case:
		return GroupParticipantRemoved{Participant: handleFromProto(ev.GetParticipantRemoved().GetParticipant())}, true
	case imessagev1.GroupChangeEvent_ParticipantLeft_case:
		return GroupParticipantLeft{Participant: handleFromProto(ev.GetParticipantLeft().GetParticipant())}, true
	case imessagev1.GroupChangeEvent_IconChanged_case:
		return GroupIconChanged{}, true
	case imessagev1.GroupChangeEvent_IconRemoved_case:
		return GroupIconRemoved{}, true
	default:
		return nil, false
	}
}

func groupEventFromChange(ev *imessagev1.GroupChangeEvent, sequence uint64) (GroupEvent, bool) {
	if ev == nil {
		return nil, false
	}
	change, ok := groupChangeFromProto(ev)
	if !ok {
		return nil, false
	}
	return GroupChanged{
		EventContext: EventContext{
			Actor:      handlePtrFromProto(ev.GetActor()),
			ChatGUID:   ev.GetChatGuid(),
			IsFromMe:   ev.GetIsFromMe(),
			OccurredAt: timeFromProto(ev.GetOccurredAt()),
		},
		Sequence: sequence,
		Change:   change,
	}, true
}

func groupEventFromProto(pb *imessagev1.SubscribeGroupEventsResponse) (GroupEvent, bool, error) {
	if pb.WhichPayload() != imessagev1.SubscribeGroupEventsResponse_GroupChanged_case {
		return nil, false, nil
	}
	ev, ok := groupEventFromChange(pb.GetGroupChanged(), pb.GetSequence())
	return ev, ok, nil
}
