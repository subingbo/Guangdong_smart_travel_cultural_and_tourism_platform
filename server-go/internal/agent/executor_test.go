package agent

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

func TestSpawnExecutorRejectsWhenRunningAndQueueAreFull(t *testing.T) {
	e := NewSpawnExecutor(2, 2, time.Second)
	release := make(chan struct{})

	var started int32
	run := func(ctx context.Context) {
		atomic.AddInt32(&started, 1)
		<-release
	}
	for i := 0; i < 4; i++ {
		if err := e.Submit(context.Background(), fmt.Sprintf("t%d", i), run, nil); err != nil {
			t.Fatalf("submit %d: %v", i, err)
		}
	}
	if err := e.Submit(context.Background(), "overflow", run, nil); !errors.Is(err, ErrQueueFull) {
		t.Fatalf("expected ErrQueueFull, got %v", err)
	}
	close(release)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	e.Shutdown(ctx)
}

func TestSpawnExecutorQueueTimeout(t *testing.T) {
	e := NewSpawnExecutor(1, 1, 40*time.Millisecond)
	release := make(chan struct{})
	var started int32
	if err := e.Submit(context.Background(), "first", func(ctx context.Context) {
		atomic.AddInt32(&started, 1)
		<-release
	}, nil); err != nil {
		t.Fatal(err)
	}
	waitStarted(t, &started, 1)

	timedOut := make(chan struct{}, 1)
	if err := e.Submit(context.Background(), "second", func(ctx context.Context) {
		t.Error("timed-out task must not run")
	}, func() {
		timedOut <- struct{}{}
	}); err != nil {
		t.Fatal(err)
	}
	select {
	case <-timedOut:
	case <-time.After(2 * time.Second):
		t.Fatal("queue timeout did not fire")
	}
	close(release)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	e.Shutdown(ctx)
}

func TestSpawnExecutorRemovesQueuedTask(t *testing.T) {
	e := NewSpawnExecutor(1, 1, time.Second)
	release := make(chan struct{})
	var started int32
	if err := e.Submit(context.Background(), "first", func(ctx context.Context) {
		atomic.AddInt32(&started, 1)
		<-release
	}, nil); err != nil {
		t.Fatal(err)
	}
	waitStarted(t, &started, 1)

	var ran int32
	if err := e.Submit(context.Background(), "second", func(ctx context.Context) {
		atomic.StoreInt32(&ran, 1)
	}, func() {
		t.Error("removed task must not time out")
	}); err != nil {
		t.Fatal(err)
	}
	if !e.Remove("second") {
		t.Fatal("expected to remove a queued task")
	}
	if e.Remove("first") {
		t.Fatal("a running task must not be removable")
	}
	close(release)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	e.Shutdown(ctx)
	if atomic.LoadInt32(&ran) != 0 {
		t.Fatal("removed task ran")
	}
}

func TestFixedExecutorRejectsWhenQueueIsFull(t *testing.T) {
	e := NewFixedExecutor(1, 1, time.Second)
	release := make(chan struct{})
	var started int32
	if err := e.Submit(context.Background(), "first", func(ctx context.Context) {
		atomic.AddInt32(&started, 1)
		<-release
	}, nil); err != nil {
		t.Fatal(err)
	}
	waitStarted(t, &started, 1)

	if err := e.Submit(context.Background(), "second", func(ctx context.Context) {}, nil); err != nil {
		t.Fatalf("second should queue: %v", err)
	}
	if err := e.Submit(context.Background(), "third", func(ctx context.Context) {}, nil); !errors.Is(err, ErrQueueFull) {
		t.Fatalf("expected ErrQueueFull, got %v", err)
	}
	close(release)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	e.Shutdown(ctx)
}

func TestFixedExecutorQueueTimeout(t *testing.T) {
	e := NewFixedExecutor(1, 1, 40*time.Millisecond)
	release := make(chan struct{})
	var started int32
	if err := e.Submit(context.Background(), "first", func(ctx context.Context) {
		atomic.AddInt32(&started, 1)
		<-release
	}, nil); err != nil {
		t.Fatal(err)
	}
	waitStarted(t, &started, 1)

	timedOut := make(chan struct{}, 1)
	if err := e.Submit(context.Background(), "second", func(ctx context.Context) {
		t.Error("timed-out task must not run")
	}, func() {
		timedOut <- struct{}{}
	}); err != nil {
		t.Fatal(err)
	}
	select {
	case <-timedOut:
	case <-time.After(2 * time.Second):
		t.Fatal("queue timeout did not fire")
	}
	close(release)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	e.Shutdown(ctx)
}

func waitStarted(t *testing.T, counter *int32, want int32) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if atomic.LoadInt32(counter) >= want {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("task did not start")
}
