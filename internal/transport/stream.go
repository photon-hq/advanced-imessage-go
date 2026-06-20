package transport

import (
	"context"
	"errors"
	"iter"
	"sync"
	"sync/atomic"

	"connectrpc.com/connect"
)

// errAlreadyConsumed is returned via Err when Events is called more than once.
var errAlreadyConsumed = errors.New("imessage: stream already consumed")

// MapFunc converts a protobuf frame of type F into a domain event of type T.
// It returns (event, true, nil) to emit, (zero, false, nil) to drop the frame
// (for example heartbeats or unrecognized payloads), or (zero, false, err) to
// fail the stream.
type MapFunc[F, T any] func(*F) (T, bool, error)

// OpenFunc opens the underlying server-streaming RPC against ctx.
type OpenFunc[F any] func(ctx context.Context) (*connect.ServerStreamForClient[F], error)

// Subscription is a single-consumer view over a server-streaming RPC. Iterate
// it with Events, check Err after the loop, and always Close it.
type Subscription[T any] struct {
	cancel    context.CancelFunc
	out       chan result[T]
	done      chan struct{}
	consumed  atomic.Bool
	closeOnce sync.Once

	mu  sync.Mutex
	err error
}

type result[T any] struct {
	val T
	err error
}

// Subscribe starts a subscription. A goroutine immediately begins receiving
// from the stream opened by open and mapping frames with mapFn. The goroutine
// is bound to a derived context and is guaranteed to exit on Close or when the
// parent ctx is cancelled.
func Subscribe[F, T any](ctx context.Context, open OpenFunc[F], mapFn MapFunc[F, T]) *Subscription[T] {
	ctx, cancel := context.WithCancel(ctx)
	s := &Subscription[T]{
		cancel: cancel,
		out:    make(chan result[T]),
		done:   make(chan struct{}),
	}
	go pump(ctx, s, open, mapFn)
	return s
}

func pump[F, T any](ctx context.Context, s *Subscription[T], open OpenFunc[F], mapFn MapFunc[F, T]) {
	// Defers run last-in-first-out: s.out is closed first (so a consumer
	// ranging over it observes the end), then s.done is closed (so a blocked
	// Close returns only after the pump has fully unwound).
	defer close(s.done)
	defer close(s.out)

	stream, err := open(ctx)
	if err != nil {
		s.trySend(ctx, result[T]{err: err})
		return
	}
	defer stream.Close()

	for {
		if !stream.Receive() {
			// A nil Err is a clean end-of-stream (io.EOF); leave Err unset.
			if e := stream.Err(); e != nil {
				s.trySend(ctx, result[T]{err: e})
			}
			return
		}
		val, emit, mErr := mapFn(stream.Msg())
		if mErr != nil {
			s.trySend(ctx, result[T]{err: mErr})
			return
		}
		if !emit {
			continue
		}
		if !s.trySend(ctx, result[T]{val: val}) {
			return
		}
	}
}

// trySend delivers r unless ctx is cancelled first, guaranteeing the pump can
// always make progress toward exit even if the consumer has stopped reading.
func (s *Subscription[T]) trySend(ctx context.Context, r result[T]) bool {
	select {
	case s.out <- r:
		return true
	case <-ctx.Done():
		return false
	}
}

// Events returns an iterator over the stream's events. It may be ranged over
// only once; a second call yields nothing and sets Err. Breaking out of the
// range loop (or letting it finish) releases the underlying RPC.
func (s *Subscription[T]) Events() iter.Seq[T] {
	return func(yield func(T) bool) {
		if !s.consumed.CompareAndSwap(false, true) {
			s.setErr(errAlreadyConsumed)
			return
		}
		defer s.Close()
		for r := range s.out {
			if r.err != nil {
				s.setErr(r.err)
				return
			}
			if !yield(r.val) {
				return
			}
		}
	}
}

// Err returns the terminal error that stopped iteration, or nil if the stream
// ended cleanly. Call it after ranging over Events.
func (s *Subscription[T]) Err() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.err
}

func (s *Subscription[T]) setErr(err error) {
	s.mu.Lock()
	if s.err == nil {
		s.err = err
	}
	s.mu.Unlock()
}

// Close cancels the RPC and blocks until the receiving goroutine has fully
// exited, so teardown is deterministic. It is safe to call multiple times.
func (s *Subscription[T]) Close() error {
	s.closeOnce.Do(func() {
		s.cancel()
		<-s.done
	})
	return nil
}
