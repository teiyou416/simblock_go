package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
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
	NetworkProfile     network.Profile
}

type Stats struct {
	AcceptedBlocks       int     `json:"accepted_blocks"`
	ObservedBlocks       int     `json:"observed_blocks"`
	MeanPropagationDelay float64 `json:"mean_propagation_delay"`
	OrphanBlocks         int     `json:"orphan_blocks"`
	OrphanRate           float64 `json:"orphan_rate"`
	TotalEvents          int     `json:"total_events"`
	AddNodeEvents        int     `json:"add_node_events"`
	AddLinkEvents        int     `json:"add_link_events"`
	AddBlockEvents       int     `json:"add_block_events"`
	FlowBlockEvents      int     `json:"flow_block_events"`
	SimulationEndEvents  int     `json:"simulation_end_events"`
	SimulationEndTime    int64   `json:"simulation_end_time"`
}

type chainTreeSnapshot struct {
	GeneratedAt time.Time       `json:"generated_at"`
	BestTipID   *uint64         `json:"best_tip_id,omitempty"`
	RootIDs     []uint64        `json:"root_ids"`
	Nodes       []chainTreeNode `json:"nodes"`
	Edges       []chainTreeEdge `json:"edges"`
}

type chainTreeNode struct {
	BlockID     uint64       `json:"block_id"`
	ParentID    *uint64      `json:"parent_id"`
	Height      uint64       `json:"height"`
	MinterID    int          `json:"minter_id"`
	Timestamp   core.SimTime `json:"timestamp"`
	OnBestChain bool         `json:"on_best_chain"`
}

type chainTreeEdge struct {
	ParentID uint64 `json:"parent_id"`
	ChildID  uint64 `json:"child_id"`
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
	if cfg.NetworkProfile.Name == "" {
		cfg.NetworkProfile = network.Bitcoin2019Profile
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
		n.SetBlockSize(s.cfg.BlockSize)
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
	regionList := s.makeRandomListFollowDistribution(s.cfg.NetworkProfile.RegionDistribution, false)
	degreeList := s.makeRandomListFollowDistribution(s.cfg.NetworkProfile.DegreeDistribution, true)
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
		n.SetBlockSize(s.cfg.BlockSize)
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

	mainChainSet := canonicalChainSet(s.bestTip())
	mainChainBlocks := 0
	for id := range mainChainSet {
		if _, seen := s.seenBlocks[id]; seen {
			mainChainBlocks++
		}
	}

	total := len(s.seenBlocks)
	orphans := total - mainChainBlocks
	if orphans < 0 {
		orphans = 0
	}
	orphanRate := 0.0
	if total > 0 {
		orphanRate = float64(orphans) / float64(total)
	}

	eventCounts := make(map[string]int)
	var simulationEndTime int64
	for _, event := range s.events {
		kind, _ := event["kind"].(string)
		eventCounts[kind]++
		if kind == "simulation-end" {
			if content, ok := event["content"].(map[string]any); ok {
				switch timestamp := content["timestamp"].(type) {
				case core.SimTime:
					simulationEndTime = int64(timestamp)
				case int64:
					simulationEndTime = timestamp
				case int:
					simulationEndTime = int64(timestamp)
				}
			}
		}
	}

	return Stats{
		AcceptedBlocks:       s.acceptedBlocks,
		ObservedBlocks:       len(s.seenBlocks),
		MeanPropagationDelay: mean,
		OrphanBlocks:         orphans,
		OrphanRate:           orphanRate,
		TotalEvents:          len(s.events),
		AddNodeEvents:        eventCounts["add-node"],
		AddLinkEvents:        eventCounts["add-link"],
		AddBlockEvents:       eventCounts["add-block"],
		FlowBlockEvents:      eventCounts["flow-block"],
		SimulationEndEvents:  eventCounts["simulation-end"],
		SimulationEndTime:    simulationEndTime,
	}
}

func (s *Simulator) bestTip() *core.Block {
	var best *core.Block
	var bestWork uint64
	for _, n := range s.nodes {
		tip := n.Tip()
		if tip == nil {
			continue
		}
		work := chainWork(tip)
		if best == nil ||
			work > bestWork ||
			(work == bestWork && tip.Height() > best.Height()) ||
			(work == bestWork && tip.Height() == best.Height() && tip.ID() < best.ID()) {
			best = tip
			bestWork = work
		}
	}
	return best
}

func chainWork(block *core.Block) uint64 {
	if data, ok := consensus.PoWDataFromBlock(block); ok {
		return data.TotalDifficulty
	}
	return block.Height()
}

func canonicalChainSet(tip *core.Block) map[uint64]struct{} {
	chain := make(map[uint64]struct{})
	for b := tip; b != nil; b = b.Parent() {
		chain[b.ID()] = struct{}{}
	}
	return chain
}

func (s *Simulator) writeOutputs(stats Stats) error {
	if err := os.MkdirAll(s.cfg.OutputDir, 0o755); err != nil {
		return err
	}

	outputPath := filepath.Join(s.cfg.OutputDir, "output.txt")
	if err := writeTextFile(outputPath, buildEventsText(s.events)); err != nil {
		return err
	}

	staticPath := filepath.Join(s.cfg.OutputDir, "static.txt")
	if err := writeTextFile(staticPath, buildStaticText()); err != nil {
		return err
	}

	metricsPath := filepath.Join(s.cfg.OutputDir, "metrics.txt")
	if err := writeTextFile(metricsPath, buildMetricsText(stats)); err != nil {
		return err
	}

	chainTreeTextPath := filepath.Join(s.cfg.OutputDir, "chain_tree.txt")
	if err := writeTextFile(chainTreeTextPath, buildChainTreeText(s.buildChainTreeSnapshot())); err != nil {
		return err
	}

	return nil
}

func (s *Simulator) buildChainTreeSnapshot() chainTreeSnapshot {
	ids := make([]uint64, 0, len(s.seenBlocks))
	for id := range s.seenBlocks {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	bestChain := canonicalChainSet(s.bestTip())
	nodes := make([]chainTreeNode, 0, len(ids))
	edges := make([]chainTreeEdge, 0, len(ids))
	rootIDs := make([]uint64, 0, 1)

	for _, id := range ids {
		b := s.seenBlocks[id]
		var parentID *uint64
		if pid, ok := b.ParentID(); ok {
			pidCopy := pid
			parentID = &pidCopy
			edges = append(edges, chainTreeEdge{
				ParentID: pid,
				ChildID:  id,
			})
		} else {
			rootIDs = append(rootIDs, id)
		}
		_, onBestChain := bestChain[id]
		nodes = append(nodes, chainTreeNode{
			BlockID:     id,
			ParentID:    parentID,
			Height:      b.Height(),
			MinterID:    b.MinterID(),
			Timestamp:   b.Time(),
			OnBestChain: onBestChain,
		})
	}

	sort.Slice(edges, func(i, j int) bool {
		if edges[i].ParentID != edges[j].ParentID {
			return edges[i].ParentID < edges[j].ParentID
		}
		return edges[i].ChildID < edges[j].ChildID
	})
	sort.Slice(rootIDs, func(i, j int) bool { return rootIDs[i] < rootIDs[j] })

	snapshot := chainTreeSnapshot{
		GeneratedAt: time.Now().UTC(),
		RootIDs:     rootIDs,
		Nodes:       nodes,
		Edges:       edges,
	}
	if bestTip := s.bestTip(); bestTip != nil {
		bestTipID := bestTip.ID()
		snapshot.BestTipID = &bestTipID
	}
	return snapshot
}

func writeTextFile(path, body string) error {
	return os.WriteFile(path, []byte(body), 0o644)
}

func buildEventsText(events []map[string]any) string {
	var b strings.Builder
	for i, event := range events {
		kind, _ := event["kind"].(string)
		fmt.Fprintf(&b, "[%d] kind=%s", i, kind)
		content, _ := event["content"].(map[string]any)
		if len(content) > 0 {
			keys := make([]string, 0, len(content))
			for k := range content {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				fmt.Fprintf(&b, " %s=%v", k, content[k])
			}
		}
		b.WriteByte('\n')
	}
	return b.String()
}

func buildStaticText() string {
	var b strings.Builder
	b.WriteString("region_id\tname\n")
	for _, r := range network.DefaultRegions {
		fmt.Fprintf(&b, "%d\t%s\n", r.ID, r.Name)
	}
	return b.String()
}

func buildMetricsText(stats Stats) string {
	var b strings.Builder
	fmt.Fprintf(&b, "generated_at: %s\n", time.Now().UTC().Format(time.RFC3339))
	fmt.Fprintf(&b, "accepted_blocks: %d\n", stats.AcceptedBlocks)
	fmt.Fprintf(&b, "observed_blocks: %d\n", stats.ObservedBlocks)
	fmt.Fprintf(&b, "mean_propagation_delay: %.6f\n", stats.MeanPropagationDelay)
	fmt.Fprintf(&b, "orphan_blocks: %d\n", stats.OrphanBlocks)
	fmt.Fprintf(&b, "orphan_rate: %.6f\n", stats.OrphanRate)
	fmt.Fprintf(&b, "total_events: %d\n", stats.TotalEvents)
	fmt.Fprintf(&b, "add_node_events: %d\n", stats.AddNodeEvents)
	fmt.Fprintf(&b, "add_link_events: %d\n", stats.AddLinkEvents)
	fmt.Fprintf(&b, "add_block_events: %d\n", stats.AddBlockEvents)
	fmt.Fprintf(&b, "flow_block_events: %d\n", stats.FlowBlockEvents)
	fmt.Fprintf(&b, "simulation_end_events: %d\n", stats.SimulationEndEvents)
	fmt.Fprintf(&b, "simulation_end_time: %d\n", stats.SimulationEndTime)
	return b.String()
}

func buildChainTreeText(snapshot chainTreeSnapshot) string {
	var b strings.Builder
	fmt.Fprintf(&b, "generated_at: %s\n", snapshot.GeneratedAt.Format(time.RFC3339))
	if snapshot.BestTipID != nil {
		fmt.Fprintf(&b, "best_tip_id: %d\n", *snapshot.BestTipID)
	} else {
		b.WriteString("best_tip_id: none\n")
	}
	fmt.Fprintf(&b, "nodes: %d\n", len(snapshot.Nodes))
	fmt.Fprintf(&b, "edges: %d\n", len(snapshot.Edges))
	b.WriteString("tree_compressed:\n")

	children := make(map[uint64][]uint64, len(snapshot.Nodes))
	nodeByID := make(map[uint64]chainTreeNode, len(snapshot.Nodes))
	for _, n := range snapshot.Nodes {
		nodeByID[n.BlockID] = n
	}
	for _, e := range snapshot.Edges {
		children[e.ParentID] = append(children[e.ParentID], e.ChildID)
	}
	for id := range children {
		sort.Slice(children[id], func(i, j int) bool { return children[id][i] < children[id][j] })
	}

	var walk func(id uint64, prefix string, isLast bool)
	walk = func(id uint64, prefix string, isLast bool) {
		start, ok := nodeByID[id]
		if !ok {
			return
		}

		branch := "+- "
		nextPrefix := prefix + "|  "
		if isLast {
			branch = "\\- "
			nextPrefix = prefix + "   "
		}

		end := start
		allBest := start.OnBestChain
		for len(children[end.BlockID]) == 1 {
			nextID := children[end.BlockID][0]
			nextNode, exists := nodeByID[nextID]
			if !exists {
				break
			}
			end = nextNode
			allBest = allBest && nextNode.OnBestChain
			if len(children[end.BlockID]) != 1 {
				break
			}
		}

		bestMark := ""
		if allBest {
			bestMark = " *best"
		}
		if start.BlockID == end.BlockID {
			fmt.Fprintf(&b, "%s%sblock=%d h=%d minter=%d ts=%d%s\n",
				prefix, branch, start.BlockID, start.Height, start.MinterID, start.Timestamp, bestMark)
		} else {
			fmt.Fprintf(&b, "%s%spath %d(h=%d)->%d(h=%d) len=%d%s\n",
				prefix, branch,
				start.BlockID, start.Height,
				end.BlockID, end.Height,
				end.Height-start.Height+1,
				bestMark,
			)
		}

		kids := children[end.BlockID]
		for i, childID := range kids {
			walk(childID, nextPrefix, i == len(kids)-1)
		}
	}

	for i, rootID := range snapshot.RootIDs {
		walk(rootID, "", i == len(snapshot.RootIDs)-1)
	}
	return b.String()
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
