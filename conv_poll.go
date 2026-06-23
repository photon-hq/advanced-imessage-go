package imessage

import (
	"cmp"
	"slices"

	imessagev1 "buf.build/gen/go/photon-hq/imessage/protocolbuffers/go/photon/imessage/v1"
)

func pollFromProto(pb *imessagev1.PollInfo) Poll {
	if pb == nil {
		return Poll{}
	}
	p := Poll{
		PollMessageGUID: pb.GetPollMessageGuid(),
		ChatGUID:        pb.GetChatGuid(),
		Title:           pb.GetTitle(),
	}
	// Options are kept in server definition order (see [Poll.Options]).
	for _, o := range pb.GetOptions() {
		p.Options = append(p.Options, pollOptionFromProto(o))
	}
	for _, v := range pb.GetVotes() {
		p.Votes = append(p.Votes, PollParticipantVote{
			Participant: handleFromProto(v.GetParticipant()),
			OptionID:    v.GetOptionIdentifier(),
		})
	}
	// Deterministic order (participant address, then option id); see [Poll.Votes].
	slices.SortFunc(p.Votes, func(a, b PollParticipantVote) int {
		if c := cmp.Compare(a.Participant.Address, b.Participant.Address); c != 0 {
			return c
		}
		return cmp.Compare(a.OptionID, b.OptionID)
	})
	return p
}

func pollOptionFromProto(pb *imessagev1.PollOption) PollOption {
	return PollOption{
		OptionID:      pb.GetOptionIdentifier(),
		Text:          pb.GetText(),
		CreatorHandle: pb.GetCreatorHandle(),
	}
}

func pollOptionsFromProto(in []*imessagev1.PollOption) []PollOption {
	if len(in) == 0 {
		return nil
	}
	out := make([]PollOption, 0, len(in))
	for _, o := range in {
		out = append(out, pollOptionFromProto(o))
	}
	return out
}

func pollDeltaFromProto(ev *imessagev1.PollChangeEvent) (PollChangeDelta, bool) {
	switch ev.WhichChange() {
	case imessagev1.PollChangeEvent_Created_case:
		c := ev.GetCreated()
		return PollCreated{Title: c.GetTitle(), Options: pollOptionsFromProto(c.GetOptions())}, true
	case imessagev1.PollChangeEvent_OptionAdded_case:
		c := ev.GetOptionAdded()
		return PollOptionAdded{Title: c.GetTitle(), Options: pollOptionsFromProto(c.GetOptions())}, true
	case imessagev1.PollChangeEvent_Voted_case:
		return PollVoted{OptionID: ev.GetVoted().GetOptionIdentifier()}, true
	case imessagev1.PollChangeEvent_Unvoted_case:
		return PollUnvoted{OptionID: ev.GetUnvoted().GetOptionIdentifier()}, true
	default:
		return nil, false
	}
}

func pollEventFromChange(ev *imessagev1.PollChangeEvent, sequence uint64) (PollEvent, bool) {
	if ev == nil {
		return nil, false
	}
	delta, ok := pollDeltaFromProto(ev)
	if !ok {
		return nil, false
	}
	return PollChanged{
		EventContext: EventContext{
			Actor:      handlePtrFromProto(ev.GetActor()),
			ChatGUID:   ev.GetChatGuid(),
			IsFromMe:   ev.GetIsFromMe(),
			OccurredAt: timeFromProto(ev.GetOccurredAt()),
		},
		Sequence:        sequence,
		PollMessageGUID: ev.GetPollMessageGuid(),
		Delta:           delta,
	}, true
}

func pollEventFromProto(pb *imessagev1.SubscribePollEventsResponse) (PollEvent, bool, error) {
	if pb.WhichPayload() != imessagev1.SubscribePollEventsResponse_PollChanged_case {
		return nil, false, nil
	}
	ev, ok := pollEventFromChange(pb.GetPollChanged(), pb.GetSequence())
	return ev, ok, nil
}
