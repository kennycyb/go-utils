package future

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestStartFuture_Success(t *testing.T) {
	ctx := context.Background()
	fut := StartFuture(ctx, func(ctx context.Context) (string, error) {
		return "success", nil
	})

	val, err := fut.Await(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if val != "success" {
		t.Fatalf("expected 'success', got %v", val)
	}
}

func TestStartFuture_Error(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("test error")
	fut := StartFuture(ctx, func(ctx context.Context) (string, error) {
		return "", expectedErr
	})

	val, err := fut.Await(ctx)
	if err != expectedErr {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
	if val != "" {
		t.Fatalf("expected empty string, got %v", val)
	}
}

func TestStartFuture_Panic(t *testing.T) {
	ctx := context.Background()
	fut := StartFuture(ctx, func(ctx context.Context) (string, error) {
		panic("test panic")
	})

	val, err := fut.Await(ctx)
	if err == nil {
		t.Fatal("expected error due to panic")
	}
	if !strings.Contains(err.Error(), "panic: test panic") {
		t.Fatalf("expected panic error, got %v", err)
	}
	if val != "" {
		t.Fatalf("expected empty string, got %v", val)
	}
}

func TestStartFuture_CanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	fut := StartFuture(ctx, func(ctx context.Context) (string, error) {
		t.Fatal("function should not be called")
		return "should not happen", nil
	})

	val, err := fut.Await(context.Background())
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if val != "" {
		t.Fatalf("expected empty string, got %v", val)
	}
}

func TestAwait_Timeout(t *testing.T) {
	ctx := context.Background()
	fut := StartFuture(ctx, func(ctx context.Context) (string, error) {
		time.Sleep(100 * time.Millisecond)
		return "done", nil
	})

	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()

	val, err := fut.Await(timeoutCtx)
	if err != context.DeadlineExceeded {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}
	if val != "" {
		t.Fatalf("expected empty string, got %v", val)
	}
}

func TestTry_NotReady(t *testing.T) {
	ctx := context.Background()
	fut := StartFuture(ctx, func(ctx context.Context) (string, error) {
		time.Sleep(100 * time.Millisecond)
		return "done", nil
	})

	val, err, ok := fut.Try()
	if ok {
		t.Fatal("expected not ready")
	}
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if val != "" {
		t.Fatalf("expected empty string, got %v", val)
	}
}

func TestTry_Ready(t *testing.T) {
	ctx := context.Background()
	fut := StartFuture(ctx, func(ctx context.Context) (string, error) {
		return "ready", nil
	})

	time.Sleep(10 * time.Millisecond) // give time for completion

	val, err, ok := fut.Try()
	if !ok {
		t.Fatal("expected ready")
	}
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if val != "ready" {
		t.Fatalf("expected 'ready', got %v", val)
	}
}

func TestAll_Success(t *testing.T) {
	ctx := context.Background()
	futures := []*Future[string]{
		StartFuture(ctx, func(ctx context.Context) (string, error) { return "a", nil }),
		StartFuture(ctx, func(ctx context.Context) (string, error) { return "b", nil }),
		StartFuture(ctx, func(ctx context.Context) (string, error) { return "c", nil }),
	}

	results, err := All(ctx, futures)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	expected := []string{"a", "b", "c"}
	for i, v := range results {
		if v != expected[i] {
			t.Fatalf("expected %v, got %v", expected[i], v)
		}
	}
}

func TestAll_Error(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("test error")
	futures := []*Future[string]{
		StartFuture(ctx, func(ctx context.Context) (string, error) { return "a", nil }),
		StartFuture(ctx, func(ctx context.Context) (string, error) { return "", expectedErr }),
		StartFuture(ctx, func(ctx context.Context) (string, error) { return "c", nil }),
	}

	results, err := All(ctx, futures)
	if err != expectedErr {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
	if results != nil {
		t.Fatalf("expected nil results, got %v", results)
	}
}

func TestAny_FirstCompletes(t *testing.T) {
	ctx := context.Background()
	futures := []*Future[string]{
		StartFuture(ctx, func(ctx context.Context) (string, error) {
			time.Sleep(100 * time.Millisecond)
			return "slow", nil
		}),
		StartFuture(ctx, func(ctx context.Context) (string, error) {
			time.Sleep(10 * time.Millisecond)
			return "fast", nil
		}),
		StartFuture(ctx, func(ctx context.Context) (string, error) {
			time.Sleep(50 * time.Millisecond)
			return "medium", nil
		}),
	}

	val, err, idx := Any(ctx, futures)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if val != "fast" {
		t.Fatalf("expected 'fast', got %v", val)
	}
	if idx != 1 {
		t.Fatalf("expected index 1, got %d", idx)
	}
}

func TestAny_Error(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("test error")
	futures := []*Future[string]{
		StartFuture(ctx, func(ctx context.Context) (string, error) {
			time.Sleep(100 * time.Millisecond)
			return "slow", nil
		}),
		StartFuture(ctx, func(ctx context.Context) (string, error) {
			time.Sleep(10 * time.Millisecond)
			return "", expectedErr
		}),
	}

	val, err, idx := Any(ctx, futures)
	if err != expectedErr {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
	if val != "" {
		t.Fatalf("expected empty string, got %v", val)
	}
	if idx != 1 {
		t.Fatalf("expected index 1, got %d", idx)
	}
}

func TestAny_Timeout(t *testing.T) {
	ctx := context.Background()
	futures := []*Future[string]{
		StartFuture(ctx, func(ctx context.Context) (string, error) {
			time.Sleep(100 * time.Millisecond)
			return "slow", nil
		}),
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()

	val, err, idx := Any(timeoutCtx, futures)
	if err != context.DeadlineExceeded {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}
	if val != "" {
		t.Fatalf("expected empty string, got %v", val)
	}
	if idx != -1 {
		t.Fatalf("expected index -1, got %d", idx)
	}
}