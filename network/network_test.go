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

	if got, want := m.GetBandwidth(0, 1), uint64(50); got != want {
		t.Fatalf("GetBandwidth(0,1): got=%d want=%d", got, want)
	}
}

func TestTransferTime(t *testing.T) {
	m := NewModel(
		[][]core.SimTime{
			{1, 5},
			{5, 1},
		},
		[]uint64{100_000, 30_000},
		[]uint64{80_000, 50_000},
	)

	// bottleneck bw: min(100000, 50000) = 50000 bit/s => 50 bit/ms
	// payload: 1000 bytes => 8000 bits => 160 ms
	if got, want := m.TransferTime(1000, 0, 1), core.SimTime(160); got != want {
		t.Fatalf("TransferTime(): got=%d want=%d", got, want)
	}
}
