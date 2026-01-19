package future

import (
	"context"
	"fmt"
	"runtime/debug"
)

// Result wraps a value or an error.
type Result[T any] struct {
	Value T
	Err   error
}

// Future represents a one-shot computation that will produce exactly one Result.
type Future[T any] struct {
	ch <-chan Result[T]
}

// StartFuture launches fn in a goroutine and returns a Future.
// Behavior:
//   - If ctx is already canceled, fn is not called and ctx.Err() is returned.
//   - If fn panics, the panic is converted to an error (with stack trace).
//   - Exactly one Result is sent, then the channel is closed.
//   - The send is non-blocking (buffered channel of size 1), so the goroutine never gets stuck.
func StartFuture[T any](ctx context.Context, fn func(context.Context) (T, error)) *Future[T] {
	ch := make(chan Result[T], 1)

	go func() {
		var out Result[T]

		// Ensure we always deliver exactly one result and then close.
		defer func() {
			ch <- out
			close(ch)
		}()

		// Convert panics to an error so Await doesn't block forever.
		defer func() {
			if r := recover(); r != nil {
				out.Err = fmt.Errorf("panic: %v\n%s", r, debug.Stack())
			}
		}()

		// If already canceled, don't call fn.
		select {
		case <-ctx.Done():
			var zero T
			out = Result[T]{Value: zero, Err: ctx.Err()}
			return
		default:
		}

		v, err := fn(ctx)
		out = Result[T]{Value: v, Err: err}
	}()

	return &Future[T]{ch: ch}
}

// Await blocks until the Future completes or ctx is done.
func (f *Future[T]) Await(ctx context.Context) (T, error) {
	var zero T
	select {
	case r := <-f.ch:
		return r.Value, r.Err
	case <-ctx.Done():
		return zero, ctx.Err()
	}
}

// Try performs a non-blocking check for completion.
// ok == false means the future is not ready yet.
func (f *Future[T]) Try() (value T, err error, ok bool) {
	var zero T
	select {
	case r := <-f.ch:
		return r.Value, r.Err, true
	default:
		return zero, nil, false
	}
}

// All waits for all futures to complete, returning the values or the first error.
func All[T any](ctx context.Context, futures []*Future[T]) ([]T, error) {
	out := make([]T, len(futures))
	for i, f := range futures {
		v, err := f.Await(ctx)
		if err != nil {
			return nil, err
		}
		out[i] = v
	}
	return out, nil
}

// Any returns the first completed future (its value, error, and index).
// Cancels waiting for the rest once one finishes or ctx is done.
func Any[T any](ctx context.Context, futures []*Future[T]) (T, error, int) {
	type pair struct {
		idx int
		r   Result[T]
	}
	out := make(chan pair, 1)

	innerCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	for i, f := range futures {
		i := i
		f := f
		go func() {
			v, err := f.Await(innerCtx)
			select {
			case out <- pair{i, Result[T]{v, err}}:
			case <-innerCtx.Done():
			}
		}()
	}

	select {
	case <-ctx.Done():
		var zero T
		return zero, ctx.Err(), -1
	case p := <-out:
		// Stop other waiters as soon as one completes.
		cancel()
		return p.r.Value, p.r.Err, p.idx
	}
}