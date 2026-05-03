package engine

import (
	"sort"
	"testing"

	"github.com/teiyou416/simblock_go/core"
)

type stubTask struct {
	interval core.SimTime
	run      func()
}

func (s stubTask) Interval() core.SimTime { return s.interval }
func (s stubTask) Run()                   { s.run() }

func TestTimerOrdersByTimestampThenIdentity(t *testing.T) {
	timer := NewTimer()
	var executed []string

	mk := func(name string) core.Task {
		return &stubTask{
			run: func() {
				executed = append(executed, name)
			},
		}
	}

	// Same timestamp, order is decided by pseudo identity hash.
	timer.PutTaskAt(mk("first"), 100)
	timer.PutTaskAt(mk("second"), 100)
	timer.PutTaskAt(mk("third"), 100)

	timer.RunUntilEmpty()

	entries := []struct {
		name string
		seq  uint64
	}{
		{name: "first", seq: 0},
		{name: "second", seq: 1},
		{name: "third", seq: 2},
	}
	sort.Slice(entries, func(i, j int) bool {
		ha := pseudoIdentityHash(entries[i].seq)
		hb := pseudoIdentityHash(entries[j].seq)
		if ha != hb {
			return ha < hb
		}
		return entries[i].seq < entries[j].seq
	})
	want := []string{entries[0].name, entries[1].name, entries[2].name}
	if len(executed) != len(want) {
		t.Fatalf("unexpected execution count: got=%d want=%d", len(executed), len(want))
	}
	for i := range want {
		if executed[i] != want[i] {
			t.Fatalf("unexpected execution order at index %d: got=%s want=%s", i, executed[i], want[i])
		}
	}
}

func TestTimerPutTaskUsesRelativeIntervalAndAdvancesTime(t *testing.T) {
	timer := NewTimer()
	var executed []string

	appendName := func(name string) func() {
		return func() { executed = append(executed, name) }
	}

	// CurrentTime = 0
	timer.PutTask(stubTask{interval: 10, run: appendName("t10-a")})
	timer.PutTask(stubTask{interval: 5, run: appendName("t5")})
	timer.PutTask(stubTask{interval: 10, run: appendName("t10-b")})

	if got := timer.GetTask(); got == nil {
		t.Fatal("GetTask() returned nil for non-empty queue")
	}

	timer.RunUntilEmpty()

	wantOrder := []string{"t5", "t10-a", "t10-b"}
	if pseudoIdentityHash(2) < pseudoIdentityHash(0) {
		wantOrder = []string{"t5", "t10-b", "t10-a"}
	}
	if len(executed) != len(wantOrder) {
		t.Fatalf("unexpected execution count: got=%d want=%d", len(executed), len(wantOrder))
	}
	for i := range wantOrder {
		if executed[i] != wantOrder[i] {
			t.Fatalf("unexpected execution order at index %d: got=%s want=%s", i, executed[i], wantOrder[i])
		}
	}

	if got, want := timer.CurrentTime(), core.SimTime(10); got != want {
		t.Fatalf("unexpected current time after run: got=%d want=%d", got, want)
	}
}

func TestTimerRemoveTask(t *testing.T) {
	timer := NewTimer()
	var executed []string

	taskA := &stubTask{
		interval: 1,
		run:      func() { executed = append(executed, "A") },
	}
	taskB := &stubTask{
		interval: 2,
		run:      func() { executed = append(executed, "B") },
	}

	timer.PutTask(taskA)
	timer.PutTask(taskB)

	if ok := timer.RemoveTask(taskA); !ok {
		t.Fatal("RemoveTask(taskA) = false, want true")
	}
	if ok := timer.RemoveTask(taskA); ok {
		t.Fatal("RemoveTask(taskA) = true on second call, want false")
	}

	timer.RunUntilEmpty()

	if len(executed) != 1 || executed[0] != "B" {
		t.Fatalf("unexpected executed tasks: got=%v want=[B]", executed)
	}
}
