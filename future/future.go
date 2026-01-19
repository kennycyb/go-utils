package future

import (
	"context"
	"fmt"
	"reflect"
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
// Uses reflect.Select to efficiently wait on multiple channels without spawning additional goroutines.
func Any[T any](ctx context.Context, futures []*Future[T]) (T, error, int) {
	var zero T

	if len(futures) == 0 {
		return zero, nil, -1
	}

	// Build a list of cases for reflect.Select
	// Case 0: ctx.Done()
	// Case 1..N: futures[i].ch
	cases := make([]reflect.SelectCase, len(futures)+1)
	cases[0] = reflect.SelectCase{
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(ctx.Done()),
	}
	for i, f := range futures {
		cases[i+1] = reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(f.ch),
		}
	}

	// Wait for the first case to complete
	chosen, value, ok := reflect.Select(cases)

	// Case 0: context was canceled
	if chosen == 0 {
		return zero, ctx.Err(), -1
	}

	// Case 1..N: one of the futures completed
	idx := chosen - 1
	if !ok {
		// Channel was closed without sending a value (shouldn't happen with our implementation)
		return zero, fmt.Errorf("future channel closed unexpectedly"), idx
	}

	// Extract the Result from the received value
	result := value.Interface().(Result[T])
	return result.Value, result.Err, idx
}
