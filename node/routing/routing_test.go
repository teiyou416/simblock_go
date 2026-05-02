package routing

import "testing"

type stubNode struct {
	id int
}

func (n stubNode) ID() int { return n.id }

func TestBaseTableDefaultsAndSetters(t *testing.T) {
	self := stubNode{id: 7}
	table := NewBaseTable(self)

	if got, want := table.Self().ID(), 7; got != want {
		t.Fatalf("Self().ID(): got=%d want=%d", got, want)
	}
	if got, want := table.NumConnection(), 8; got != want {
		t.Fatalf("default NumConnection(): got=%d want=%d", got, want)
	}

	table.SetNumConnection(16)
	if got, want := table.NumConnection(), 16; got != want {
		t.Fatalf("NumConnection() after set: got=%d want=%d", got, want)
	}

	table.SetNumConnection(-10)
	if got, want := table.NumConnection(), 0; got != want {
		t.Fatalf("NumConnection() negative clamp: got=%d want=%d", got, want)
	}
}

func TestBaseTableDefaultHooks(t *testing.T) {
	table := NewBaseTable(stubNode{id: 1})
	peer := stubNode{id: 2}

	if table.AddInbound(peer) {
		t.Fatal("AddInbound() default should be false")
	}
	if table.RemoveInbound(peer) {
		t.Fatal("RemoveInbound() default should be false")
	}

	// Should not panic.
	table.AcceptBlock()
}
