package network

import (
	"testing"

	"github.com/teiyou416/simblock_go/core"
)

func TestLatencyAndBandwidth(t *testing.T) {
	m := NewModel(
		[][]core.SimTime{
			{1, 5},
			{5, 1},
		},
		[]uint64{100, 30},
		[]uint64{80, 50},
	)

	if got, want := m.Latency(0, 1), core.SimTime(5); got != want {
		t.Fatalf("Latency(0,1): got=%d want=%d", got, want)
	}

	if got, want := m.Bandwidth(0, 1), uint64(50); got != want {
		t.Fatalf("Bandwidth(0,1): got=%d want=%d", got, want)
	}
}
