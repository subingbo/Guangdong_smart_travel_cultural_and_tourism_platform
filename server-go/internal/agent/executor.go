package agent

import (
	"context"
	"errors"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrQueueFull      = errors.New("agent: run queue is full")
	ErrExecutorClosed = errors.New("agent: executor is closed")
)

// Executor is the admission-control seam. SpawnExecutor is the production
// default (one goroutine per run + running/admission semaphores); FixedExecutor
// is the control arm for benchmarks and a fallback for tiny memory limits.
type Executor interface {
	Submit(ctx context.Context, id string, run func(context.Context), onAdmissionTimeout func()) error
	Remove(id string) bool
	QueuedCount() int
	RunningCount() int
	Mode() string
	Shutdown(ctx context.Context)
}

type taskState struct {
	id       string
	started  atomic.Bool
	done     atomic.Bool
	canceled atomic.Bool
	cancelCh chan struct{}
	once     sync.Once
}

func newTaskState(id string) *taskState {
	return &taskState{id: id, cancelCh: make(chan struct{})}
}

func (t *taskState) cancel() {
	t.once.Do(func() {
		t.canceled.Store(true)
		close(t.cancelCh)
	})
}

// SpawnExecutor gives every task its own goroutine. Two semaphores keep the
// running limit and the queue capacity hard-bounded.
type SpawnExecutor struct {
	running      chan struct{}
	admission    chan struct{}
	queueTimeout time.Duration

	mu     sync.Mutex
	tasks  map[string]*taskState
	closed bool
	wg     sync.WaitGroup
}

func NewSpawnExecutor(maxConcurrent, queueCapacity int, queueTimeout time.Duration) *SpawnExecutor {
	if maxConcurrent < 1 {
		maxConcurrent = 1
	}
	if queueCapacity < 0 {
		queueCapacity = 0
	}
	if queueTimeout <= 0 {
		queueTimeout = 30 * time.Second
	}
	return &SpawnExecutor{
		running:      make(chan struct{}, maxConcurrent),
		admission:    make(chan struct{}, queueCapacity),
		queueTimeout: queueTimeout,
		tasks:        make(map[string]*taskState),
	}
}

func (e *SpawnExecutor) Submit(ctx context.Context, id string, run func(context.Context), onAdmissionTimeout func()) error {
	e.mu.Lock()
	if e.closed {
		e.mu.Unlock()
		return ErrExecutorClosed
	}
	task := newTaskState(id)
	admitted := false
	select {
	case e.running <- struct{}{}:
		admitted = true
	default:
	}
	if !admitted {
		select {
		case e.admission <- struct{}{}:
		default:
			e.mu.Unlock()
			return ErrQueueFull
		}
	} else {
		// The task already owns a running permit; Reset cannot cancel it.
		task.started.Store(true)
	}
	e.tasks[id] = task
	e.mu.Unlock()

	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		defer e.finish(id)

		if admitted {
			e.runSafely(ctx, task, run)
			<-e.running
			return
		}

		timer := time.NewTimer(e.queueTimeout)
		defer timer.Stop()
		select {
		case e.running <- struct{}{}:
			<-e.admission
			if task.canceled.Load() {
				<-e.running
				return
			}
			task.started.Store(true)
			e.runSafely(ctx, task, run)
			<-e.running
		case <-task.cancelCh:
			<-e.admission
		case <-ctx.Done():
			<-e.admission
		case <-timer.C:
			<-e.admission
			if onAdmissionTimeout != nil {
				onAdmissionTimeout()
			}
		}
	}()
	return nil
}

func (e *SpawnExecutor) Remove(id string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	task, ok := e.tasks[id]
	if !ok || task.started.Load() || task.done.Load() {
		return false
	}
	task.cancel()
	return true
}

func (e *SpawnExecutor) QueuedCount() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	n := 0
	for _, task := range e.tasks {
		if !task.started.Load() && !task.done.Load() {
			n++
		}
	}
	return n
}

func (e *SpawnExecutor) RunningCount() int { return len(e.running) }

func (e *SpawnExecutor) Mode() string { return "spawn" }

func (e *SpawnExecutor) Shutdown(ctx context.Context) {
	e.mu.Lock()
	e.closed = true
	for _, task := range e.tasks {
		if !task.started.Load() {
			task.cancel()
		}
	}
	e.mu.Unlock()

	done := make(chan struct{})
	go func() {
		e.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-ctx.Done():
	}
}

func (e *SpawnExecutor) finish(id string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if task, ok := e.tasks[id]; ok {
		task.done.Store(true)
		delete(e.tasks, id)
	}
}

func (e *SpawnExecutor) runSafely(ctx context.Context, task *taskState, run func(context.Context)) {
	defer func() {
		if p := recover(); p != nil {
			log.Printf("agent: run %s panicked: %v", task.id, p)
		}
	}()
	run(ctx)
}

// FixedExecutor runs N workers over a bounded queue. It is the benchmark
// control arm and the low-memory fallback.
type FixedExecutor struct {
	queue        chan *fixedTask
	queueTimeout time.Duration
	workers      int

	mu     sync.Mutex
	tasks  map[string]*taskState
	closed bool
	done   chan struct{}
	wg     sync.WaitGroup
}

type fixedTask struct {
	task      *taskState
	ctx       context.Context
	run       func(context.Context)
	onTimeout func()
	timer     *time.Timer
	timedOut  atomic.Bool
}

func NewFixedExecutor(maxConcurrent, queueCapacity int, queueTimeout time.Duration) *FixedExecutor {
	if maxConcurrent < 1 {
		maxConcurrent = 1
	}
	if queueCapacity < 1 {
		queueCapacity = 1
	}
	if queueTimeout <= 0 {
		queueTimeout = 30 * time.Second
	}
	e := &FixedExecutor{
		queue:        make(chan *fixedTask, queueCapacity),
		queueTimeout: queueTimeout,
		workers:      maxConcurrent,
		tasks:        make(map[string]*taskState),
		done:         make(chan struct{}),
	}
	e.wg.Add(maxConcurrent)
	for i := 0; i < maxConcurrent; i++ {
		go e.worker()
	}
	return e
}

func (e *FixedExecutor) Submit(ctx context.Context, id string, run func(context.Context), onAdmissionTimeout func()) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return ErrExecutorClosed
	}
	task := newTaskState(id)
	ft := &fixedTask{task: task, ctx: ctx, run: run, onTimeout: onAdmissionTimeout}
	// Start the timer before publishing the task to the queue: a worker may
	// dequeue it immediately, and reading ft.timer must not race this write.
	ft.timer = time.AfterFunc(e.queueTimeout, func() {
		if ft.task.started.Load() || ft.task.done.Load() || ft.task.canceled.Load() {
			return
		}
		if ft.timedOut.CompareAndSwap(false, true) {
			ft.task.cancel()
			if ft.onTimeout != nil {
				ft.onTimeout()
			}
		}
	})
	select {
	case e.queue <- ft:
		e.tasks[id] = task
		return nil
	default:
		ft.timer.Stop()
		return ErrQueueFull
	}
}

func (e *FixedExecutor) worker() {
	defer e.wg.Done()
	for {
		select {
		case <-e.done:
			return
		case ft := <-e.queue:
			if ft == nil {
				continue
			}
			if ft.timer != nil {
				ft.timer.Stop()
			}
			if ft.timedOut.Load() || ft.task.canceled.Load() {
				e.finish(ft.task.id)
				continue
			}
			ft.task.started.Store(true)
			e.runSafely(ft.task, ft.ctx, ft.run)
			e.finish(ft.task.id)
		}
	}
}

func (e *FixedExecutor) runSafely(task *taskState, ctx context.Context, run func(context.Context)) {
	defer func() {
		if p := recover(); p != nil {
			log.Printf("agent: run %s panicked: %v", task.id, p)
		}
	}()
	run(ctx)
}

func (e *FixedExecutor) Remove(id string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	task, ok := e.tasks[id]
	if !ok || task.started.Load() || task.done.Load() {
		return false
	}
	task.cancel()
	return true
}

func (e *FixedExecutor) QueuedCount() int { return len(e.queue) }

func (e *FixedExecutor) RunningCount() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	n := 0
	for _, task := range e.tasks {
		if task.started.Load() && !task.done.Load() {
			n++
		}
	}
	return n
}

func (e *FixedExecutor) Mode() string { return "fixed" }

func (e *FixedExecutor) Shutdown(ctx context.Context) {
	e.mu.Lock()
	if !e.closed {
		e.closed = true
		close(e.done)
	}
	e.mu.Unlock()

	for {
		select {
		case ft := <-e.queue:
			if ft != nil {
				if ft.timer != nil {
					ft.timer.Stop()
				}
				ft.task.cancel()
				e.finish(ft.task.id)
			}
		default:
			goto wait
		}
	}
wait:
	done := make(chan struct{})
	go func() {
		e.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-ctx.Done():
	}
}

func (e *FixedExecutor) finish(id string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if task, ok := e.tasks[id]; ok {
		task.done.Store(true)
		delete(e.tasks, id)
	}
}
