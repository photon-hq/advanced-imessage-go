package imessage

import (
	imessagev1 "buf.build/gen/go/photon-hq/imessage/protocolbuffers/go/photon/imessage/v1"
)

// catchUpEventFromProto dispatches a unified catch-up frame to the appropriate
// family change-mapper. Every live event variant also satisfies [CatchUpEvent].
func catchUpEventFromProto(pb *imessagev1.CatchUpEventsResponse) (CatchUpEvent, bool, error) {
	seq := pb.GetSequence()
	switch pb.WhichPayload() {
	case imessagev1.CatchUpEventsResponse_MessageChanged_case:
		ev, ok := messageEventFromChange(pb.GetMessageChanged(), seq)
		return asCatchUp(ev, ok)
	case imessagev1.CatchUpEventsResponse_ChatChanged_case:
		ev, ok := chatEventFromChange(pb.GetChatChanged(), seq)
		return asCatchUp(ev, ok)
	case imessagev1.CatchUpEventsResponse_GroupChanged_case:
		ev, ok := groupEventFromChange(pb.GetGroupChanged(), seq)
		return asCatchUp(ev, ok)
	case imessagev1.CatchUpEventsResponse_PollChanged_case:
		ev, ok := pollEventFromChange(pb.GetPollChanged(), seq)
		return asCatchUp(ev, ok)
	case imessagev1.CatchUpEventsResponse_Complete_case:
		return CatchupComplete{HeadSequence: pb.GetComplete().GetHeadSequence()}, true, nil
	default:
		return nil, false, nil
	}
}

// asCatchUp narrows a family event interface to [CatchUpEvent]. The assertion
// always succeeds for real variants because each implements isCatchUpEvent.
func asCatchUp(ev any, ok bool) (CatchUpEvent, bool, error) {
	if !ok {
		return nil, false, nil
	}
	cu, isCatchUp := ev.(CatchUpEvent)
	if !isCatchUp {
		return nil, false, nil
	}
	return cu, true, nil
}
