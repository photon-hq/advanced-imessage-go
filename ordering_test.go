package imessage

import (
	"slices"
	"testing"
	"time"

	imessagev1 "buf.build/gen/go/photon-hq/imessage/protocolbuffers/go/photon/imessage/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestAppliedReactionsSortedOldestFirst(t *testing.T) {
	t0 := time.Date(2026, 6, 20, 12, 0, 0, 0, time.UTC)
	mk := func(guid string, at time.Time) *imessagev1.MessageAppliedReaction {
		r := &imessagev1.MessageAppliedReaction{}
		r.SetMessageGuid(guid)
		r.SetDateCreated(timestamppb.New(at))
		return r
	}
	// Deliberately out of order; "a" and "b-early" tie on time and must break by GUID.
	in := []*imessagev1.MessageAppliedReaction{
		mk("c", t0.Add(2*time.Minute)),
		mk("b-late", t0.Add(3*time.Minute)),
		mk("b-early", t0.Add(1*time.Minute)),
		mk("a", t0.Add(1*time.Minute)),
	}
	got := make([]string, 0, len(in))
	for _, r := range appliedReactionsFromProto(in) {
		got = append(got, r.MessageGUID)
	}
	if want := []string{"a", "b-early", "c", "b-late"}; !slices.Equal(got, want) {
		t.Errorf("order = %v, want %v", got, want)
	}
}

func TestPlacedStickersSortedOldestFirst(t *testing.T) {
	t0 := time.Date(2026, 6, 20, 12, 0, 0, 0, time.UTC)
	mk := func(guid string, at time.Time) *imessagev1.MessagePlacedSticker {
		s := &imessagev1.MessagePlacedSticker{}
		s.SetMessageGuid(guid)
		s.SetDateCreated(timestamppb.New(at))
		return s
	}
	in := []*imessagev1.MessagePlacedSticker{
		mk("z", t0.Add(5*time.Minute)),
		mk("y", t0.Add(1*time.Minute)),
	}
	got := make([]string, 0, len(in))
	for _, s := range placedStickersFromProto(in) {
		got = append(got, s.MessageGUID)
	}
	if want := []string{"y", "z"}; !slices.Equal(got, want) {
		t.Errorf("order = %v, want %v", got, want)
	}
}

func TestPollVotesSortedAndOptionsPreserved(t *testing.T) {
	mkVote := func(addr, opt string) *imessagev1.PollParticipantVote {
		p := &imessagev1.SingleServiceAddressInfo{}
		p.SetAddress(addr)
		v := &imessagev1.PollParticipantVote{}
		v.SetParticipant(p)
		v.SetOptionIdentifier(opt)
		return v
	}
	mkOption := func(id string) *imessagev1.PollOption {
		o := &imessagev1.PollOption{}
		o.SetOptionIdentifier(id)
		return o
	}
	pb := &imessagev1.PollInfo{}
	// Options must stay in server definition order (NOT sorted).
	pb.SetOptions([]*imessagev1.PollOption{mkOption("opt-z"), mkOption("opt-a")})
	// Votes are sorted by (participant address, option id).
	pb.SetVotes([]*imessagev1.PollParticipantVote{
		mkVote("bob", "opt-z"),
		mkVote("alice", "opt-z"),
		mkVote("alice", "opt-a"),
	})

	got := pollFromProto(pb)

	gotOpts := make([]string, 0, len(got.Options))
	for _, o := range got.Options {
		gotOpts = append(gotOpts, o.OptionID)
	}
	if want := []string{"opt-z", "opt-a"}; !slices.Equal(gotOpts, want) {
		t.Errorf("options = %v, want definition order %v", gotOpts, want)
	}

	gotVotes := make([]string, 0, len(got.Votes))
	for _, v := range got.Votes {
		gotVotes = append(gotVotes, v.Participant.Address+"/"+v.OptionID)
	}
	if want := []string{"alice/opt-a", "alice/opt-z", "bob/opt-z"}; !slices.Equal(gotVotes, want) {
		t.Errorf("votes = %v, want %v", gotVotes, want)
	}
}
