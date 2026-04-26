package engine

import (
	"container/heap"
	"reflect"

	"github.com/teiyou416/simblock_go/core"
)

type scheduledTask struct {
	task      core.Task
	timestamp core.SimTime
	sequence  uint64
	index     int
}

type taskMinHeap []*scheduledTask

func (h taskMinHeap) Len() int { return len(h) }

func (h taskMinHeap) Less(i, j int) bool {
	if h[i].timestamp != h[j].timestamp {
		return h[i].timestamp < h[j].timestamp
	}
	return h[i].sequence < h[j].sequence
}

func (h taskMinHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *taskMinHeap) Push(x any) {
	item := x.(*scheduledTask)
	item.index = len(*h)
	*h = append(*h, item)
}

func (h *taskMinHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*h = old[:n-1]
	return item
}

// Timer is a deterministic global event scheduler.
//
// Ordering rule:
// 1) smaller timestamp first
// 2) for equal timestamps, smaller sequence first (FIFO for same time)
type Timer struct {
	currentTime core.SimTime
	nextSeq     uint64
	pq          taskMinHeap
	taskMap     map[uintptr]*scheduledTask
}

func NewTimer() *Timer {
	t := &Timer{
		taskMap: make(map[uintptr]*scheduledTask),
	}
	heap.Init(&t.pq)
	return t
}

func taskIdentity(task core.Task) (uintptr, bool) {
	v := reflect.ValueOf(task)
	if !v.IsValid() {
		return 0, false
	}
	if v.Kind() != reflect.Pointer || v.IsNil() {
		return 0, false
	}
	return v.Pointer(), true
}

func (t *Timer) CurrentTime() core.SimTime {
	return t.currentTime
}

func (t *Timer) Len() int {
	return t.pq.Len()
}

func (t *Timer) HasTask() bool {
	return t.Len() > 0
}

// PutTask schedules a task at CurrentTime + task.Interval().
func (t *Timer) PutTask(task core.Task) {
	t.PutTaskAt(task, t.currentTime+task.Interval())
}

// PutTaskAt schedules a task at an absolute simulation timestamp.
func (t *Timer) PutTaskAt(task core.Task, timestamp core.SimTime) {
	// Keep one queued entry per task instance.
	taskID, trackable := taskIdentity(task)
	if trackable {
		if old, ok := t.taskMap[taskID]; ok && old.index >= 0 {
			heap.Remove(&t.pq, old.index)
			delete(t.taskMap, taskID)
		}
	}

	entry := &scheduledTask{
		task:      task,
		timestamp: timestamp,
		sequence:  t.nextSeq,
		index:     -1,
	}
	t.nextSeq++
	heap.Push(&t.pq, entry)
	if trackable {
		t.taskMap[taskID] = entry
	}
}

// GetTask returns the next task to execute without removing it.
// It returns nil if the queue is empty.
func (t *Timer) GetTask() core.Task {
	if t.pq.Len() == 0 {
		return nil
	}
	return t.pq[0].task
}

// RunTask pops the next task, advances simulation time, and executes the task.
// It returns false when the queue is empty.
func (t *Timer) RunTask() bool {
	if t.pq.Len() == 0 {
		return false
	}

	next := heap.Pop(&t.pq).(*scheduledTask)
	if taskID, ok := taskIdentity(next.task); ok {
		delete(t.taskMap, taskID)
	}
	t.currentTime = next.timestamp
	next.task.Run()
	return true
}

// RemoveTask removes a queued task. It returns true when the task existed.
func (t *Timer) RemoveTask(task core.Task) bool {
	taskID, trackable := taskIdentity(task)
	if !trackable {
		return false
	}

	entry, ok := t.taskMap[taskID]
	if !ok || entry.index < 0 {
		return false
	}

	heap.Remove(&t.pq, entry.index)
	delete(t.taskMap, taskID)
	return true
}

// RunUntilEmpty is the core simulation loop for pure event-driven execution.
func (t *Timer) RunUntilEmpty() {
	for t.RunTask() {
	}
}
