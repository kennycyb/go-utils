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

	// Use Await to ensure the future completes
	_, _ = fut.Await(ctx)

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

func TestAwait_MultipleCalls(t *testing.T) {
	ctx := context.Background()
	fut := StartFuture(ctx, func(ctx context.Context) (string, error) {
		return "result", nil
	})

	// First call should work
	val1, err1 := fut.Await(ctx)
	if err1 != nil {
		t.Fatalf("first Await: expected no error, got %v", err1)
	}
	if val1 != "result" {
		t.Fatalf("first Await: expected 'result', got %v", val1)
	}

	// Second call should return the same result
	val2, err2 := fut.Await(ctx)
	if err2 != nil {
		t.Fatalf("second Await: expected no error, got %v", err2)
	}
	if val2 != "result" {
		t.Fatalf("second Await: expected 'result', got %v", val2)
	}

	// Third call should also work
	val3, err3 := fut.Await(ctx)
	if err3 != nil {
		t.Fatalf("third Await: expected no error, got %v", err3)
	}
	if val3 != "result" {
		t.Fatalf("third Await: expected 'result', got %v", val3)
	}
}

func TestAwait_MultipleCallsWithError(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("test error")
	fut := StartFuture(ctx, func(ctx context.Context) (string, error) {
		return "", expectedErr
	})

	// First call should return the error
	val1, err1 := fut.Await(ctx)
	if err1 != expectedErr {
		t.Fatalf("first Await: expected error %v, got %v", expectedErr, err1)
	}
	if val1 != "" {
		t.Fatalf("first Await: expected empty string, got %v", val1)
	}

	// Second call should return the same error
	val2, err2 := fut.Await(ctx)
	if err2 != expectedErr {
		t.Fatalf("second Await: expected error %v, got %v", expectedErr, err2)
	}
	if val2 != "" {
		t.Fatalf("second Await: expected empty string, got %v", val2)
	}
}

func TestTry_MultipleCalls(t *testing.T) {
	ctx := context.Background()
	fut := StartFuture(ctx, func(ctx context.Context) (string, error) {
		return "result", nil
	})

	// Use Await to ensure the future completes
	_, _ = fut.Await(ctx)

	// First call should succeed
	val1, err1, ok1 := fut.Try()
	if !ok1 {
		t.Fatal("first Try: expected ready")
	}
	if err1 != nil {
		t.Fatalf("first Try: expected no error, got %v", err1)
	}
	if val1 != "result" {
		t.Fatalf("first Try: expected 'result', got %v", val1)
	}

	// Second call should return the same result
	val2, err2, ok2 := fut.Try()
	if !ok2 {
		t.Fatal("second Try: expected ready")
	}
	if err2 != nil {
		t.Fatalf("second Try: expected no error, got %v", err2)
	}
	if val2 != "result" {
		t.Fatalf("second Try: expected 'result', got %v", val2)
	}

	// Third call should also work
	val3, err3, ok3 := fut.Try()
	if !ok3 {
		t.Fatal("third Try: expected ready")
	}
	if err3 != nil {
		t.Fatalf("third Try: expected no error, got %v", err3)
	}
	if val3 != "result" {
		t.Fatalf("third Try: expected 'result', got %v", val3)
	}
}

func TestTry_MultipleCallsWithError(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("test error")
	fut := StartFuture(ctx, func(ctx context.Context) (string, error) {
		return "", expectedErr
	})

	// Use Await to ensure the future completes
	_, _ = fut.Await(ctx)

	// First call should return the error
	val1, err1, ok1 := fut.Try()
	if !ok1 {
		t.Fatal("first Try: expected ready")
	}
	if err1 != expectedErr {
		t.Fatalf("first Try: expected error %v, got %v", expectedErr, err1)
	}
	if val1 != "" {
		t.Fatalf("first Try: expected empty string, got %v", val1)
	}

	// Second call should return the same error
	val2, err2, ok2 := fut.Try()
	if !ok2 {
		t.Fatal("second Try: expected ready")
	}
	if err2 != expectedErr {
		t.Fatalf("second Try: expected error %v, got %v", expectedErr, err2)
	}
	if val2 != "" {
		t.Fatalf("second Try: expected empty string, got %v", val2)
	}
}

func TestAwait_ConcurrentCalls(t *testing.T) {
	ctx := context.Background()
	
	// Start multiple goroutines that all call Await
	const numGoroutines = 10
	
	// Use channels to coordinate test execution
	startSignal := make(chan struct{})
	readyCount := make(chan struct{}, numGoroutines)
	
	fut := StartFuture(ctx, func(ctx context.Context) (string, error) {
		<-startSignal // Wait for all goroutines to be ready
		return "concurrent", nil
	})

	results := make(chan string, numGoroutines)
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			readyCount <- struct{}{} // Signal that this goroutine is ready
			val, err := fut.Await(ctx)
			results <- val
			errors <- err
		}()
	}

	// Wait for all goroutines to be ready
	for i := 0; i < numGoroutines; i++ {
		<-readyCount
	}
	
	// Now signal the future to complete
	close(startSignal)

	// Collect all results
	for i := 0; i < numGoroutines; i++ {
		val := <-results
		err := <-errors
		if err != nil {
			t.Fatalf("goroutine %d: expected no error, got %v", i, err)
		}
		if val != "concurrent" {
			t.Fatalf("goroutine %d: expected 'concurrent', got %v", i, val)
		}
	}
}

func TestTry_ConcurrentCalls(t *testing.T) {
	ctx := context.Background()
	fut := StartFuture(ctx, func(ctx context.Context) (string, error) {
		return "concurrent", nil
	})

	// Use Await to ensure the future completes
	_, _ = fut.Await(ctx)

	// Start multiple goroutines that all call Try
	const numGoroutines = 10
	results := make(chan string, numGoroutines)
	errors := make(chan error, numGoroutines)
	oks := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			val, err, ok := fut.Try()
			results <- val
			errors <- err
			oks <- ok
		}()
	}

	// Collect all results
	for i := 0; i < numGoroutines; i++ {
		val := <-results
		err := <-errors
		ok := <-oks
		if !ok {
			t.Fatalf("goroutine %d: expected ready", i)
		}
		if err != nil {
			t.Fatalf("goroutine %d: expected no error, got %v", i, err)
		}
		if val != "concurrent" {
			t.Fatalf("goroutine %d: expected 'concurrent', got %v", i, val)
		}
	}
}