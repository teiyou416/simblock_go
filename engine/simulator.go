package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/teiyou416/simblock_go/core"
	"github.com/teiyou416/simblock_go/internal/javarand"
	"github.com/teiyou416/simblock_go/network"
	"github.com/teiyou416/simblock_go/node"
	"github.com/teiyou416/simblock_go/node/consensus"
	"github.com/teiyou416/simblock_go/tasks"
)

type SimulatorConfig struct {
	NumNodes           int
	TargetInterval     core.SimTime
	EndTime            core.SimTime
	EndBlockHeight     int
	BlockSize          uint64
	OutputDir          string
	RandomSeed         int64
	ConnectionsPerNode int
	JavaCompatible     bool
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
	rng     *javarand.Random

	nodes []*node.Node

	events []map[string]any

	acceptedBlocks int
	seenByBlock    map[uint64]map[int]struct{}
	seenBlocks     map[uint64]*core.Block
	delays         []core.SimTime

	initialDifficulty uint64
}

var (
	javaRegionDistribution = []float64{0.3316, 0.4998, 0.0090, 0.1177, 0.0224, 0.0195}
	javaDegreeDistribution = []float64{
		0.025, 0.050, 0.075, 0.10, 0.20, 0.30, 0.40, 0.50, 0.60, 0.70,
		0.80, 0.85, 0.90, 0.95, 0.97, 0.97, 0.98, 0.99, 0.995, 1.0,
	}
)

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
	if cfg.JavaCompatible && cfg.EndBlockHeight <= 0 {
		cfg.EndBlockHeight = 3
	}
	return &Simulator{
		cfg:         cfg,
		timer:       timer,
		network:     net,
		rng:         javarand.New(cfg.RandomSeed),
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

	if s.cfg.JavaCompatible {
		s.network.EnableJavaLatencyJitter(s.rng)
		return s.setupJavaCompatible(regionCount)
	}
	return s.setupSimple(regionCount)
}

func (s *Simulator) setupSimple(regionCount int) error {
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
		n.SetRNG(s.rng)
		n.SetConsensus(consensus.NewPoWWithSource(consensus.PoWConfig{
			InitialDifficulty: 1,
			TargetInterval:    s.cfg.TargetInterval,
			AdjustmentWindow:  1,
			DifficultyMode:    consensus.DifficultyStatic,
		}, s.rng))

		s.nodes = append(s.nodes, n)
		totalHashPower += hashPower
		s.logAddNode(n)
	}
	return s.finishSetup(totalHashPower, false)
}

func (s *Simulator) setupJavaCompatible(regionCount int) error {
	s.nodes = make([]*node.Node, 0, s.cfg.NumNodes)
	regionList := s.makeRandomListFollowDistribution(javaRegionDistribution, false)
	degreeList := s.makeRandomListFollowDistribution(javaDegreeDistribution, true)
	useCBRList := s.makeRandomBoolList(0.964)
	churnList := s.makeRandomBoolList(0.976)

	totalHashPower := uint64(0)
	for i := 0; i < s.cfg.NumNodes; i++ {
		id := i + 1
		region := regionList[i] % regionCount
		numConn := degreeList[i] + 1
		hashPower := s.genMiningPower()
		useCBR := useCBRList[i]
		isChurn := churnList[i]

		n := node.NewWithHashPower(id, region, hashPower)
		n.BindEnvironment(s.timer, s.network)
		n.SetNumConnections(numConn)
		n.SetCompactBlockRelay(useCBR)
		n.SetChurnNode(isChurn)
		if isChurn {
			n.SetCompactFailureRate(0.27)
		} else {
			n.SetCompactFailureRate(0.13)
		}
		n.SetBlockAcceptedObserver(s.onBlockAccepted)
		n.SetFlowObserver(s.onFlowBlock)
		n.SetRNG(s.rng)
		n.SetConsensus(consensus.NewPoWWithSource(consensus.PoWConfig{
			InitialDifficulty: 1,
			TargetInterval:    s.cfg.TargetInterval,
			AdjustmentWindow:  1,
			DifficultyMode:    consensus.DifficultyStatic,
		}, s.rng))

		s.nodes = append(s.nodes, n)
		totalHashPower += hashPower
		s.logAddNode(n)
	}

	for _, from := range s.nodes {
		candidates := make([]int, 0, len(s.nodes))
		for i := 0; i < len(s.nodes); i++ {
			candidates = append(candidates, i)
		}
		s.rng.Shuffle(len(candidates), func(i, j int) {
			candidates[i], candidates[j] = candidates[j], candidates[i]
		})
		for _, idx := range candidates {
			if from.OutboundCount() >= from.NumConnections() {
				break
			}
			to := s.nodes[idx]
			if from.AddNeighbor(to) {
				s.logAddLink(to, from)
				s.logAddLink(from, to)
			}
		}
	}
	return s.finishSetup(totalHashPower, true)
}

func (s *Simulator) finishSetup(totalHashPower uint64, javaMode bool) error {
	initialDiff := totalHashPower * uint64(s.cfg.TargetInterval)
	if initialDiff == 0 {
		initialDiff = 1
	}
	s.initialDifficulty = initialDiff
	for _, n := range s.nodes {
		n.SetConsensus(consensus.NewPoWWithSource(consensus.PoWConfig{
			InitialDifficulty: initialDiff,
			TargetInterval:    s.cfg.TargetInterval,
			AdjustmentWindow:  1,
			DifficultyMode:    consensus.DifficultyStatic,
		}, s.rng))
	}

	if javaMode {
		return nil
	}

	for i, from := range s.nodes {
		for step := 1; step <= s.cfg.ConnectionsPerNode; step++ {
			to := s.nodes[(i+step)%len(s.nodes)]
			if from.AddNeighbor(to) {
				s.logAddLink(from, to)
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

	currentBlockHeight := 1
	for s.timer.HasTask() {
		if s.cfg.JavaCompatible {
			if nextTask, ok := s.timer.GetTask().(*tasks.MiningTask); ok {
				if parentHeight, hasParent := nextTask.ParentHeight(); hasParent && int(parentHeight) == currentBlockHeight {
					currentBlockHeight++
				}
				if s.cfg.EndBlockHeight > 0 && currentBlockHeight > s.cfg.EndBlockHeight {
					break
				}
			}
		} else if s.cfg.EndTime > 0 && s.timer.CurrentTime() >= s.cfg.EndTime {
			break
		}

		if !s.timer.RunTask() {
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
	r := uint64(s.rng.Float64() * float64(total))
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

func (s *Simulator) makeRandomListFollowDistribution(distribution []float64, cumulative bool) []int {
	list := make([]int, 0, s.cfg.NumNodes)
	idx := 0
	if cumulative {
		for ; idx < len(distribution); idx++ {
			for len(list) <= int(float64(s.cfg.NumNodes)*distribution[idx]) {
				list = append(list, idx)
			}
		}
		for len(list) < s.cfg.NumNodes {
			list = append(list, idx)
		}
	} else {
		acc := 0.0
		for ; idx < len(distribution); idx++ {
			acc += distribution[idx]
			for len(list) <= int(float64(s.cfg.NumNodes)*acc) {
				list = append(list, idx)
			}
		}
		for len(list) < s.cfg.NumNodes {
			list = append(list, idx)
		}
	}
	s.rng.Shuffle(len(list), func(i, j int) {
		list[i], list[j] = list[j], list[i]
	})
	return list[:s.cfg.NumNodes]
}

func (s *Simulator) makeRandomBoolList(rate float64) []bool {
	list := make([]bool, s.cfg.NumNodes)
	for i := 0; i < s.cfg.NumNodes; i++ {
		list[i] = i < int(float64(s.cfg.NumNodes)*rate)
	}
	s.rng.Shuffle(len(list), func(i, j int) {
		list[i], list[j] = list[j], list[i]
	})
	return list
}

func (s *Simulator) genMiningPower() uint64 {
	v := int64(s.rng.NormFloat64()*100000 + 400000)
	if v < 1 {
		v = 1
	}
	return uint64(v)
}

func (s *Simulator) logAddNode(n *node.Node) {
	s.events = append(s.events, map[string]any{
		"kind": "add-node",
		"content": map[string]any{
			"timestamp": 0,
			"node-id":   n.ID(),
			"region-id": n.Region(),
		},
	})
}

func (s *Simulator) logAddLink(from, to *node.Node) {
	s.events = append(s.events, map[string]any{
		"kind": "add-link",
		"content": map[string]any{
			"timestamp":     s.timer.CurrentTime(),
			"begin-node-id": from.ID(),
			"end-node-id":   to.ID(),
		},
	})
}
