package consensus

import (
	"math/rand"
	"testing"

	"github.com/teiyou416/simblock_go/core"
)

func TestPoWGenesisAndBuildChildBlock(t *testing.T) {
	pow := NewPoW(PoWConfig{
		InitialDifficulty: 100,
		TargetInterval:    10,
		AdjustmentWindow:  3,
	}, rand.New(rand.NewSource(1)))

	g := pow.GenesisBlock(1)
	gd, ok := PoWDataFromBlock(g)
	if !ok {
		t.Fatal("genesis consensus data is not PoWData")
	}
	if gd.NextDifficulty != 100 {
		t.Fatalf("genesis next difficulty: got=%d want=100", gd.NextDifficulty)
	}

	b1 := pow.BuildChildBlock(g, 2, 10)
	if b1 == nil {
		t.Fatal("b1 is nil")
	}
	d1, ok := PoWDataFromBlock(b1)
	if !ok {
		t.Fatal("b1 consensus data is not PoWData")
	}
	if d1.Difficulty != 100 || d1.TotalDifficulty != 100 {
		t.Fatalf("b1 data mismatch: got=%+v", d1)
	}
}

func TestPoWReceivedBlockValidation(t *testing.T) {
	pow := NewPoW(PoWConfig{
		InitialDifficulty: 50,
		TargetInterval:    10,
		AdjustmentWindow:  2,
	}, rand.New(rand.NewSource(2)))

	g := pow.GenesisBlock(1)
	b1 := pow.BuildChildBlock(g, 2, 10)
	b2 := pow.BuildChildBlock(b1, 3, 20)

	if !pow.IsReceivedBlockValid(b1, g) {
		t.Fatal("expected b1 valid over genesis")
	}
	if !pow.IsReceivedBlockValid(b2, b1) {
		t.Fatal("expected b2 valid over b1")
	}

	// tamper total difficulty
	badData := PoWData{
		Difficulty:      50,
		TotalDifficulty: 9999,
		NextDifficulty:  50,
	}
	bad := core.NewBlock(g, 9, 11, badData)
	if pow.IsReceivedBlockValid(bad, b1) {
		t.Fatal("tampered total difficulty should be invalid")
	}
}

func TestPoWMintingReturnsTaskWithPositiveInterval(t *testing.T) {
	pow := NewPoW(PoWConfig{
		InitialDifficulty: 100,
		TargetInterval:    10,
		AdjustmentWindow:  3,
	}, rand.New(rand.NewSource(3)))

	g := pow.GenesisBlock(1)
	task := pow.Minting(g, 1, 1000)
	if task == nil {
		t.Fatal("Minting() returned nil task")
	}
	if got := task.Interval(); got <= 0 {
		t.Fatalf("interval should be positive, got=%d", got)
	}
}

func TestPoWDifficultyAdjustment(t *testing.T) {
	pow := NewPoW(PoWConfig{
		InitialDifficulty: 100,
		TargetInterval:    10,
		AdjustmentWindow:  2,
	}, rand.New(rand.NewSource(4)))

	g := pow.GenesisBlock(1)
	b1 := pow.BuildChildBlock(g, 2, 10)
	d1, _ := PoWDataFromBlock(b1)
	if d1.NextDifficulty != 100 {
		t.Fatalf("unexpected difficulty before window boundary: got=%d want=100", d1.NextDifficulty)
	}

	// Height 2 is a window boundary (AdjustmentWindow=2).
	// Parent timespan over last window: 10-0=10ms; expected is 20ms,
	// so difficulty should increase to 200.
	b2 := pow.BuildChildBlock(b1, 3, 20)
	d2, _ := PoWDataFromBlock(b2)
	if got, want := d2.NextDifficulty, uint64(200); got != want {
		t.Fatalf("retargeted next difficulty: got=%d want=%d", got, want)
	}
}
