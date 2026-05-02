package engine

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/teiyou416/simblock_go/core"
	"github.com/teiyou416/simblock_go/network"
	"github.com/teiyou416/simblock_go/node"
	"github.com/teiyou416/simblock_go/node/consensus"
)

type SimulatorConfig struct {
	NumNodes           int
	TargetInterval     core.SimTime
	EndTime            core.SimTime
	BlockSize          uint64
	OutputDir          string
	RandomSeed         int64
	ConnectionsPerNode int
}

type Stats struct {
	AcceptedBlocks       int     `json:"accepted_blocks"`
	ObservedBlocks       int     `json:"observed_blocks"`
	MeanPropagationDelay float64 `json:"mean_propagation_delay"`
	OrphanBlocks         int     `json:"orphan_blocks"`
	OrphanRate           float64 `json:"orphan_rate"`
}

type Simulator struct {
	cfg     SimulatorConfig
	timer   *Timer
	network *network.Model
	rng     *rand.Rand

	nodes []*node.Node

	events []map[string]any

	acceptedBlocks int
	seenByBlock    map[uint64]map[int]struct{}
	seenBlocks     map[uint64]*core.Block
	delays         []core.SimTime
	propagatedOnce bool

	initialDifficulty uint64
}

func NewSimulator(cfg SimulatorConfig, timer *Timer, net *network.Model) *Simulator {
	if cfg.NumNodes <= 0 {
		cfg.NumNodes = 2
	}
	if cfg.TargetInterval <= 0 {
		cfg.TargetInterval = 600_000
	}
	if cfg.OutputDir == "" {
		cfg.OutputDir = "output"
	}
	if cfg.RandomSeed == 0 {
		cfg.RandomSeed = 10
	}
	if cfg.ConnectionsPerNode <= 0 {
		cfg.ConnectionsPerNode = 8
	}
	if timer == nil {
		timer = NewTimer()
	}
	return &Simulator{
		cfg:         cfg,
		timer:       timer,
		network:     net,
		rng:         rand.New(rand.NewSource(cfg.RandomSeed)),
		seenByBlock: make(map[uint64]map[int]struct{}),
		seenBlocks:  make(map[uint64]*core.Block),
	}
}

func (s *Simulator) Nodes() []*node.Node {
	out := make([]*node.Node, len(s.nodes))
	copy(out, s.nodes)
	return out
}

func (s *Simulator) Setup() error {
	if s.network == nil {
		return fmt.Errorf("network model is nil")
	}
	regionCount := s.network.RegionCount()
	if regionCount == 0 {
		return fmt.Errorf("network has no regions")
	}

	s.nodes = make([]*node.Node, 0, s.cfg.NumNodes)
	totalHashPower := uint64(0)
	for i := 0; i < s.cfg.NumNodes; i++ {
		region := i % regionCount
		hashPower := uint64(1000 + s.rng.Intn(1000))
		n := node.NewWithHashPower(i+1, region, hashPower)
		n.BindEnvironment(s.timer, s.network)
		n.SetCompactBlockRelay(true)
		n.SetCompactFailureRate(0.13)
		n.SetBlockAcceptedObserver(s.onBlockAccepted)
		n.SetFlowObserver(s.onFlowBlock)

		pow := consensus.NewPoW(consensus.PoWConfig{
			InitialDifficulty: 1, // will be set after total hashpower computed
			TargetInterval:    s.cfg.TargetInterval,
			AdjustmentWindow:  1,
			DifficultyMode:    consensus.DifficultyStatic,
		}, rand.New(rand.NewSource(s.rng.Int63())))
		n.SetConsensus(pow)

		s.nodes = append(s.nodes, n)
		totalHashPower += hashPower

		s.events = append(s.events, map[string]any{
			"kind": "add-node",
			"content": map[string]any{
				"timestamp": 0,
				"node-id":   n.ID(),
				"region-id": n.Region(),
			},
		})
	}

	// Rebind PoW with Java-like static initial difficulty.
	initialDiff := totalHashPower * uint64(s.cfg.TargetInterval)
	if initialDiff == 0 {
		initialDiff = 1
	}
	s.initialDifficulty = initialDiff
	for _, n := range s.nodes {
		pow := consensus.NewPoW(consensus.PoWConfig{
			InitialDifficulty: initialDiff,
			TargetInterval:    s.cfg.TargetInterval,
			AdjustmentWindow:  1,
			DifficultyMode:    consensus.DifficultyStatic,
		}, rand.New(rand.NewSource(s.rng.Int63())))
		n.SetConsensus(pow)
	}

	// Deterministic directed mesh by ring expansion.
	for i, from := range s.nodes {
		for step := 1; step <= s.cfg.ConnectionsPerNode; step++ {
			to := s.nodes[(i+step)%len(s.nodes)]
			if from.AddNeighbor(to) {
				s.events = append(s.events, map[string]any{
					"kind": "add-link",
					"content": map[string]any{
						"timestamp":     0,
						"begin-node-id": from.ID(),
						"end-node-id":   to.ID(),
					},
				})
			}
		}
	}

	return nil
}

func (s *Simulator) Run() (Stats, error) {
	if len(s.nodes) == 0 {
		if err := s.Setup(); err != nil {
			return Stats{}, err
		}
	}

	minter := s.pickGenesisMinter()
	if minter == nil {
		return Stats{}, fmt.Errorf("failed to select genesis minter")
	}

	genesis := core.GenesisBlock(minter.ID(), consensus.PoWData{
		Difficulty:      0,
		TotalDifficulty: 0,
		NextDifficulty:  s.initialDifficulty,
	})
	minter.ReceiveBlock(genesis)

	guard := 0
	for s.timer.HasTask() {
		if s.cfg.EndTime > 0 && s.timer.CurrentTime() >= s.cfg.EndTime {
			break
		}
		if !s.timer.RunTask() {
			break
		}
		guard++
		if s.propagatedOnce && guard > 10 {
			break
		}
		if guard > 2_000_000 {
			break
		}
	}

	s.events = append(s.events, map[string]any{
		"kind": "simulation-end",
		"content": map[string]any{
			"timestamp": s.timer.CurrentTime(),
		},
	})

	stats := s.collectStats()
	if err := s.writeOutputs(stats); err != nil {
		return Stats{}, err
	}
	return stats, nil
}

func (s *Simulator) pickGenesisMinter() *node.Node {
	if len(s.nodes) == 0 {
		return nil
	}
	var total uint64
	for _, n := range s.nodes {
		total += n.HashPower()
	}
	if total == 0 {
		return s.nodes[0]
	}
	r := uint64(s.rng.Int63n(int64(total)))
	var cum uint64
	for _, n := range s.nodes {
		cum += n.HashPower()
		if r < cum {
			return n
		}
	}
	return s.nodes[len(s.nodes)-1]
}

func (s *Simulator) onBlockAccepted(self *node.Node, block *core.Block, timestamp core.SimTime) {
	s.acceptedBlocks++
	s.seenBlocks[block.ID()] = block

	seen := s.seenByBlock[block.ID()]
	if seen == nil {
		seen = make(map[int]struct{})
		s.seenByBlock[block.ID()] = seen
	}
	if _, exists := seen[self.ID()]; !exists {
		seen[self.ID()] = struct{}{}
		s.delays = append(s.delays, timestamp-block.Time())
	}

	if block.Height() > 0 && len(seen) >= 2 {
		s.propagatedOnce = true
	}

	s.events = append(s.events, map[string]any{
		"kind": "add-block",
		"content": map[string]any{
			"timestamp": timestamp,
			"node-id":   self.ID(),
			"block-id":  block.ID(),
		},
	})
}

func (s *Simulator) onFlowBlock(from, to *node.Node, block *core.Block, transmission, reception core.SimTime) {
	s.events = append(s.events, map[string]any{
		"kind": "flow-block",
		"content": map[string]any{
			"transmission-timestamp": transmission,
			"reception-timestamp":    reception,
			"begin-node-id":          from.ID(),
			"end-node-id":            to.ID(),
			"block-id":               block.ID(),
		},
	})
}

func (s *Simulator) collectStats() Stats {
	var mean float64
	if len(s.delays) > 0 {
		var sum core.SimTime
		for _, d := range s.delays {
			sum += d
		}
		mean = float64(sum) / float64(len(s.delays))
	}

	orphanSet := make(map[uint64]struct{})
	for _, n := range s.nodes {
		for _, b := range n.Orphans() {
			orphanSet[b.ID()] = struct{}{}
		}
	}

	mainChainBlocks := 0
	if len(s.nodes) > 0 && s.nodes[0].Tip() != nil {
		mainChainBlocks = int(s.nodes[0].Tip().Height() + 1)
	}
	orphans := len(orphanSet)
	total := mainChainBlocks + orphans
	orphanRate := 0.0
	if total > 0 {
		orphanRate = float64(orphans) / float64(total)
	}

	return Stats{
		AcceptedBlocks:       s.acceptedBlocks,
		ObservedBlocks:       len(s.seenBlocks),
		MeanPropagationDelay: mean,
		OrphanBlocks:         orphans,
		OrphanRate:           orphanRate,
	}
}

func (s *Simulator) writeOutputs(stats Stats) error {
	if err := os.MkdirAll(s.cfg.OutputDir, 0o755); err != nil {
		return err
	}

	outputPath := filepath.Join(s.cfg.OutputDir, "output.json")
	if err := writeJSONFile(outputPath, s.events); err != nil {
		return err
	}

	staticRegions := make([]map[string]any, 0, len(network.DefaultRegions))
	for _, r := range network.DefaultRegions {
		staticRegions = append(staticRegions, map[string]any{
			"id":   r.ID,
			"name": r.Name,
		})
	}
	staticPath := filepath.Join(s.cfg.OutputDir, "static.json")
	if err := writeJSONFile(staticPath, map[string]any{"region": staticRegions}); err != nil {
		return err
	}

	metricsPath := filepath.Join(s.cfg.OutputDir, "metrics.json")
	if err := writeJSONFile(metricsPath, map[string]any{
		"generated_at": time.Now().UTC().Format(time.RFC3339),
		"stats":        stats,
	}); err != nil {
		return err
	}

	return nil
}

func writeJSONFile(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}
