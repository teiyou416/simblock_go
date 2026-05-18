package consensus

import (
	"math"
	"math/rand"

	"github.com/teiyou416/simblock_go/core"
	"github.com/teiyou416/simblock_go/tasks"
)

type float64Source interface {
	Float64() float64
}

// PoWData is the consensus payload attached to each block.
type PoWData struct {
	Difficulty      uint64
	TotalDifficulty uint64
	NextDifficulty  uint64
}

type DifficultyMode string

const (
	// DifficultyStatic keeps next difficulty unchanged after genesis.
	// This matches the current Java SimBlock behavior.
	DifficultyStatic DifficultyMode = "static"
	// DifficultyDynamic retargets difficulty at AdjustmentWindow boundaries.
	DifficultyDynamic DifficultyMode = "dynamic"
)

// PoWConfig configures the PoW algorithm behavior.
type PoWConfig struct {
	InitialDifficulty uint64
	TargetInterval    core.SimTime
	AdjustmentWindow  uint64
	DifficultyMode    DifficultyMode
}

// PoW implements Algorithm for proof-of-work style chains.
type PoW struct {
	cfg PoWConfig
	rng float64Source
}

func NewPoW(cfg PoWConfig, rng *rand.Rand) *PoW {
	return NewPoWWithSource(cfg, rng)
}

func NewPoWWithSource(cfg PoWConfig, rng float64Source) *PoW {
	if cfg.InitialDifficulty == 0 {
		cfg.InitialDifficulty = 1
	}
	if cfg.TargetInterval <= 0 {
		cfg.TargetInterval = 1
	}
	if cfg.AdjustmentWindow == 0 {
		cfg.AdjustmentWindow = 1
	}
	if cfg.DifficultyMode == "" {
		cfg.DifficultyMode = DifficultyStatic
	}
	if rng == nil {
		rng = rand.New(rand.NewSource(1))
	}
	return &PoW{cfg: cfg, rng: rng}
}

// Minting estimates time-to-mine via an exponential distribution.
func (p *PoW) Minting(currentTip *core.Block, _ int, selfHashPower uint64) core.Task {
	if currentTip == nil {
		return nil
	}
	if selfHashPower == 0 {
		selfHashPower = 1
	}
	parentData, ok := PoWDataFromBlock(currentTip)
	if !ok {
		return nil
	}

	difficulty := parentData.NextDifficulty
	if difficulty == 0 {
		difficulty = p.cfg.InitialDifficulty
	}

	u := p.rng.Float64()
	interval := core.SimTime(-math.Log(1-u) * float64(difficulty) / float64(selfHashPower))
	if interval < 0 {
		interval = 0
	}
	return tasks.NewMiningTask(interval, nil)
}

func (p *PoW) IsReceivedBlockValid(receivedBlock, currentTip *core.Block) bool {
	if receivedBlock == nil {
		return false
	}
	receivedData, ok := PoWDataFromBlock(receivedBlock)
	if !ok {
		return false
	}

	// Genesis block is always accepted.
	parent := receivedBlock.Parent()
	if parent == nil {
		return true
	}

	parentData, ok := PoWDataFromBlock(parent)
	if !ok {
		return false
	}
	if receivedData.Difficulty < parentData.NextDifficulty {
		return false
	}
	if receivedData.TotalDifficulty != parentData.TotalDifficulty+receivedData.Difficulty {
		return false
	}

	if currentTip == nil {
		return true
	}
	currentData, ok := PoWDataFromBlock(currentTip)
	if !ok {
		return false
	}
	return receivedData.TotalDifficulty > currentData.TotalDifficulty
}

func (p *PoW) GenesisBlock(minterID int) *core.Block {
	data := PoWData{
		Difficulty:      0,
		TotalDifficulty: 0,
		NextDifficulty:  p.cfg.InitialDifficulty,
	}
	return core.GenesisBlock(minterID, data)
}

// BuildChildBlock creates a PoW block extending parent at time now.
func (p *PoW) BuildChildBlock(parent *core.Block, minterID int, now core.SimTime) *core.Block {
	return p.BuildChildBlockWithUncles(parent, minterID, now, nil)
}

func (p *PoW) BuildChildBlockWithUncles(parent *core.Block, minterID int, now core.SimTime, uncles []*core.Block) *core.Block {
	if parent == nil {
		return p.GenesisBlock(minterID)
	}
	parentData, ok := PoWDataFromBlock(parent)
	if !ok {
		return nil
	}
	diff := parentData.NextDifficulty
	if diff == 0 {
		diff = p.cfg.InitialDifficulty
	}
	total := parentData.TotalDifficulty + diff
	next := p.nextDifficulty(parent, diff)
	data := PoWData{
		Difficulty:      diff,
		TotalDifficulty: total,
		NextDifficulty:  next,
	}
	return core.NewBlockWithUncles(parent, minterID, now, data, uncles)
}

func (p *PoW) nextDifficulty(parent *core.Block, current uint64) uint64 {
	if p.cfg.DifficultyMode != DifficultyDynamic {
		return current
	}

	height := parent.Height() + 1
	if height == 0 || height%p.cfg.AdjustmentWindow != 0 {
		return current
	}

	anchorHeight := height - p.cfg.AdjustmentWindow
	anchor := parent.BlockWithHeight(anchorHeight)
	if anchor == nil {
		return current
	}

	actual := parent.Time() - anchor.Time()
	expected := core.SimTime(int64(p.cfg.AdjustmentWindow) * int64(p.cfg.TargetInterval))
	if actual <= 0 || expected <= 0 {
		return current
	}

	// Difficulty here is proportional to expected block interval:
	// next = current * expected / actual.
	next := (uint64(expected) * current) / uint64(actual)
	if next == 0 {
		next = 1
	}
	return next
}

func PoWDataFromBlock(b *core.Block) (PoWData, bool) {
	if b == nil {
		return PoWData{}, false
	}
	data, ok := b.ConsensusData().(PoWData)
	return data, ok
}
