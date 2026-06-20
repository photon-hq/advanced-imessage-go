package imessage

import (
	"context"
	"errors"
	"iter"
	"sync"
	"sync/atomic"
	"time"
)

// SequenceStore persists the last fully-processed event sequence so a resumable
// subscription can continue gap-free across reconnects and process restarts.
// Implementations must be safe for concurrent use; the library calls Save
// frequently and Load once at startup.
type SequenceStore interface {
	// Load returns the persisted cursor. ok is false when nothing has been
	// stored yet (start from the beginning).
	Load(ctx context.Context) (sequence uint64, ok bool, err error)
	// Save records the cursor after an event has been delivered.
	Save(ctx context.Context, sequence uint64) error
}

// ResumableOptions configures [Client.ResumableMessages].
type ResumableOptions struct {
	// Chat, when non-empty, limits the stream to one chat.
	Chat ChatGUID
	// MinBackoff and MaxBackoff bound the reconnect backoff. Zero values use
	// sensible defaults (500ms and 30s).
	MinBackoff time.Duration
	MaxBackoff time.Duration
}

const (
	defaultResumeMinBackoff = 500 * time.Millisecond
	defaultResumeMaxBackoff = 30 * time.Second
)

// errCatchUpIncomplete is returned internally when a catch-up stream ends
// without a CatchupComplete, triggering a backoff-and-retry.
var errCatchUpIncomplete = errors.New("imessage: catch-up ended before completion")

// ResumableMessages returns a message subscription that survives disconnects.
// On start (and after every disconnect) it replays missed events via
// [EventClient.CatchUp] from the cursor in store, then attaches a live
// subscription, persisting the cursor after each delivered event. Delivery is
// at-least-once: an event may repeat after a reconnect, so consumers should be
// idempotent. Always Close the returned subscription.
func (c *Client) ResumableMessages(ctx context.Context, store SequenceStore, opts *ResumableOptions) *MessageSubscription {
	minBackoff, maxBackoff := defaultResumeMinBackoff, defaultResumeMaxBackoff
	var chat ChatGUID
	if opts != nil {
		if opts.MinBackoff > 0 {
			minBackoff = opts.MinBackoff
		}
		if opts.MaxBackoff > 0 {
			maxBackoff = opts.MaxBackoff
		}
		chat = opts.Chat
	}

	ctx, cancel := context.WithCancel(ctx)
	r := &resumableMessages{
		client:     c,
		store:      store,
		chat:       chat,
		minBackoff: minBackoff,
		maxBackoff: maxBackoff,
		cancel:     cancel,
		out:        make(chan resumableResult),
		done:       make(chan struct{}),
	}
	go r.run(ctx)
	return &MessageSubscription{stream[MessageEvent]{src: r}}
}

type resumableResult struct {
	ev  MessageEvent
	err error
}

// resumableMessages implements eventSource[MessageEvent] with an internal
// catch-up -> live -> reconnect loop.
type resumableMessages struct {
	client     *Client
	store      SequenceStore
	chat       ChatGUID
	minBackoff time.Duration
	maxBackoff time.Duration

	cancel    context.CancelFunc
	out       chan resumableResult
	done      chan struct{}
	consumed  atomic.Bool
	closeOnce sync.Once

	mu  sync.Mutex
	err error
}

func (r *resumableMessages) run(ctx context.Context) {
	defer close(r.done)
	defer close(r.out)

	var cursor uint64
	if r.store != nil {
		seq, ok, err := r.store.Load(ctx)
		if err != nil {
			r.trySend(ctx, resumableResult{err: err})
			return
		}
		if ok {
			cursor = seq
		}
	}

	backoff := r.minBackoff
	for {
		if ctx.Err() != nil {
			return
		}
		if err := r.catchUp(ctx, &cursor); err != nil {
			if ctx.Err() != nil {
				return
			}
			next, ok := sleepBackoff(ctx, backoff, r.maxBackoff)
			if !ok {
				return
			}
			backoff = next
			continue
		}
		backoff = r.minBackoff // connection proven healthy
		if err := r.live(ctx, &cursor); err != nil && ctx.Err() != nil {
			return
		}
		if ctx.Err() != nil {
			return
		}
		next, ok := sleepBackoff(ctx, backoff, r.maxBackoff)
		if !ok {
			return
		}
		backoff = next
	}
}

// catchUp replays history into the consumer, advancing cursor, until the stream
// reports completion.
func (r *resumableMessages) catchUp(ctx context.Context, cursor *uint64) error {
	sub := r.client.Events().CatchUp(ctx, *cursor)
	defer func() { _ = sub.Close() }()

	for ev := range sub.Events() {
		switch e := ev.(type) {
		case CatchupComplete:
			if e.HeadSequence > *cursor {
				*cursor = e.HeadSequence
				r.saveCursor(ctx, *cursor)
			}
			return nil
		case MessageEvent:
			cg, seq := messageEventMeta(e)
			if r.chat != "" && cg != string(r.chat) {
				continue
			}
			if !r.trySend(ctx, resumableResult{ev: e}) {
				return ctx.Err()
			}
			if seq > 0 {
				*cursor = seq
				r.saveCursor(ctx, seq)
			}
		default:
			// Non-message catch-up events advance the server cursor but are
			// not delivered here; the next catch-up resumes from HeadSequence.
		}
	}
	if err := sub.Err(); err != nil {
		return err
	}
	return errCatchUpIncomplete
}

// live delivers events from a live subscription until it ends or errors.
func (r *resumableMessages) live(ctx context.Context, cursor *uint64) error {
	sub := r.client.Messages().Subscribe(ctx, &SubscribeOptions{Chat: r.chat})
	defer func() { _ = sub.Close() }()

	for ev := range sub.Events() {
		_, seq := messageEventMeta(ev)
		if !r.trySend(ctx, resumableResult{ev: ev}) {
			return ctx.Err()
		}
		if seq > 0 {
			*cursor = seq
			r.saveCursor(ctx, seq)
		}
	}
	return sub.Err()
}

func (r *resumableMessages) saveCursor(ctx context.Context, seq uint64) {
	if r.store == nil {
		return
	}
	// Best-effort: a failed Save only risks redelivery, which is already the
	// at-least-once contract.
	_ = r.store.Save(ctx, seq)
}

func (r *resumableMessages) trySend(ctx context.Context, res resumableResult) bool {
	select {
	case r.out <- res:
		return true
	case <-ctx.Done():
		return false
	}
}

// Events implements eventSource.
func (r *resumableMessages) Events() iter.Seq[MessageEvent] {
	return func(yield func(MessageEvent) bool) {
		if !r.consumed.CompareAndSwap(false, true) {
			r.setErr(errAlreadyConsumed)
			return
		}
		defer func() { _ = r.Close() }()
		for res := range r.out {
			if res.err != nil {
				r.setErr(res.err)
				return
			}
			if !yield(res.ev) {
				return
			}
		}
	}
}

// Err implements eventSource.
func (r *resumableMessages) Err() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.err
}

func (r *resumableMessages) setErr(err error) {
	r.mu.Lock()
	if r.err == nil {
		r.err = err
	}
	r.mu.Unlock()
}

// Close implements eventSource.
func (r *resumableMessages) Close() error {
	r.closeOnce.Do(func() {
		r.cancel()
		<-r.done
	})
	return nil
}

// errAlreadyConsumed mirrors the transport-level sentinel for a second Events
// call on a resumable stream.
var errAlreadyConsumed = errors.New("imessage: stream already consumed")

// sleepBackoff waits for the current backoff (or until ctx is done) and returns
// the next, doubled-and-capped backoff. ok is false if ctx was cancelled.
func sleepBackoff(ctx context.Context, backoff, maxBackoff time.Duration) (next time.Duration, ok bool) {
	timer := time.NewTimer(backoff)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return backoff, false
	case <-timer.C:
	}
	backoff *= 2
	if backoff > maxBackoff {
		backoff = maxBackoff
	}
	return backoff, true
}

func messageEventMeta(ev MessageEvent) (chatGUID string, sequence uint64) {
	switch e := ev.(type) {
	case MessageReceived:
		return e.ChatGUID, e.Sequence
	case MessageEdited:
		return e.ChatGUID, e.Sequence
	case MessageRead:
		return e.ChatGUID, e.Sequence
	case MessageUnsent:
		return e.ChatGUID, e.Sequence
	case MessageReactionAdded:
		return e.ChatGUID, e.Sequence
	case MessageReactionRemoved:
		return e.ChatGUID, e.Sequence
	case MessageStickerPlaced:
		return e.ChatGUID, e.Sequence
	default:
		return "", 0
	}
}
