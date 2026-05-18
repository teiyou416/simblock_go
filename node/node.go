package node

import (
	"math/rand"
	"sort"
	"strings"

	"github.com/teiyou416/simblock_go/core"
	"github.com/teiyou416/simblock_go/network"
	"github.com/teiyou416/simblock_go/node/consensus"
	"github.com/teiyou416/simblock_go/tasks"
)

const (
	defaultProcessingTime core.SimTime = 2
	defaultBlockSize      uint64       = 535000
	defaultCompactSize    uint64       = 18 * 1000
	maxUnclesPerBlock     int          = 2
	maxUncleGenerations   uint64       = 6
)

type ForkChoice string

const (
	ForkChoiceHeaviest ForkChoice = "heaviest"
	ForkChoiceGHOST    ForkChoice = "ghost"
)

type sizeBucket struct {
	count int
	ratio float64
}

var churnFailureBuckets = []sizeBucket{
	{546, 0.01}, {66, 0.02}, {39, 0.03}, {27, 0.04}, {21, 0.05}, {17, 0.06},
	{14, 0.07}, {12, 0.08}, {11, 0.09}, {10, 0.10}, {9, 0.11}, {8, 0.12},
	{7, 0.13}, {7, 0.14}, {6, 0.15}, {6, 0.16}, {5, 0.17}, {5, 0.18},
	{5, 0.19}, {4, 0.20}, {4, 0.21}, {4, 0.22}, {4, 0.23}, {4, 0.24},
	{3, 0.25}, {3, 0.26}, {3, 0.27}, {3, 0.28}, {3, 0.29}, {3, 0.30},
	{3, 0.31}, {3, 0.32}, {2, 0.33}, {2, 0.34}, {2, 0.35}, {2, 0.36},
	{2, 0.37}, {2, 0.38}, {2, 0.39}, {2, 0.40}, {2, 0.41}, {2, 0.42},
	{2, 0.43}, {2, 0.44}, {2, 0.45}, {2, 0.46}, {2, 0.47}, {2, 0.48},
	{1, 0.49}, {1, 0.50}, {1, 0.51}, {1, 0.52}, {1, 0.53}, {1, 0.54},
	{1, 0.55}, {1, 0.56}, {1, 0.57}, {1, 0.58}, {1, 0.59}, {1, 0.60},
	{1, 0.61}, {1, 0.62}, {1, 0.63}, {1, 0.64}, {1, 0.65}, {1, 0.66},
	{1, 0.67}, {1, 0.68}, {1, 0.69}, {1, 0.70}, {1, 0.71}, {1, 0.72},
	{1, 0.73}, {1, 0.74}, {1, 0.75}, {1, 0.76}, {1, 0.77}, {1, 0.78},
	{1, 0.79}, {1, 0.80}, {1, 0.81}, {1, 0.82}, {1, 0.83}, {1, 0.84},
	{1, 0.85}, {1, 0.86}, {1, 0.87}, {1, 0.88}, {1, 0.89}, {1, 0.90},
	{1, 0.91}, {1, 0.92}, {1, 0.93}, {1, 0.94}, {1, 0.95}, {1, 0.96},
}

type scheduler interface {
	PutTask(task core.Task)
	PutTaskAt(task core.Task, timestamp core.SimTime)
	RemoveTask(task core.Task) bool
	CurrentTime() core.SimTime
}

type blockBuilder interface {
	BuildChildBlock(parent *core.Block, minterID int, now core.SimTime) *core.Block
}

type blockBuilderWithUncles interface {
	BuildChildBlockWithUncles(parent *core.Block, minterID int, now core.SimTime, uncles []*core.Block) *core.Block
}

type randomSource interface {
	Float64() float64
	Intn(n int) int
	Shuffle(n int, swap func(i, j int))
}

// Node models a simulation node and protocol-side state transitions.
type Node struct {
	id        int
	region    int
	hashPower uint64

	tip     *core.Block
	orphans map[uint64]*core.Block

	outbound []*Node
	inbound  []*Node

	timer   scheduler
	network *network.Model

	consensus  consensus.Algorithm
	forkChoice ForkChoice

	useCompactBlockRelay bool
	churnNode            bool
	sendingBlock         bool
	messageQueue         []tasks.MessageTask
	downloadingBlocks    map[uint64]*core.Block
	mintingTask          core.Task
	numConnections       int

	processingTime core.SimTime
	blockSize      uint64
	compactSize    uint64
	cbrFailureRate float64
	rng            randomSource
	knownBlocks    map[uint64]*core.Block
	includedUncles map[uint64]struct{}

	onBlockAccepted func(self *Node, block *core.Block, timestamp core.SimTime)
	onFlowBlock     func(from, to *Node, block *core.Block, transmission, reception core.SimTime)
}

func New(id, region int) *Node {
	return NewWithHashPower(id, region, 1)
}

func NewWithHashPower(id, region int, hashPower uint64) *Node {
	if hashPower == 0 {
		hashPower = 1
	}
	return &Node{
		id:                id,
		region:            region,
		hashPower:         hashPower,
		orphans:           make(map[uint64]*core.Block),
		downloadingBlocks: make(map[uint64]*core.Block),
		processingTime:    defaultProcessingTime,
		blockSize:         defaultBlockSize,
		compactSize:       defaultCompactSize,
		numConnections:    8,
		rng:               rand.New(rand.NewSource(1)),
		forkChoice:        ForkChoiceHeaviest,
		knownBlocks:       make(map[uint64]*core.Block),
		includedUncles:    make(map[uint64]struct{}),
	}
}

func (n *Node) BindEnvironment(timer scheduler, net *network.Model) {
	n.timer = timer
	n.network = net
}

func (n *Node) SetConsensus(algo consensus.Algorithm) {
	n.consensus = algo
}

func (n *Node) SetForkChoice(choice string) {
	switch ForkChoice(strings.ToLower(choice)) {
	case ForkChoiceGHOST:
		n.forkChoice = ForkChoiceGHOST
	default:
		n.forkChoice = ForkChoiceHeaviest
	}
}

func (n *Node) ForkChoice() string {
	return string(n.forkChoice)
}

func (n *Node) SetCompactBlockRelay(enabled bool) {
	n.useCompactBlockRelay = enabled
}

func (n *Node) SetCompactFailureRate(rate float64) {
	if rate < 0 {
		rate = 0
	}
	if rate > 1 {
		rate = 1
	}
	n.cbrFailureRate = rate
}

func (n *Node) SetBlockSize(size uint64) {
	if size == 0 {
		return
	}
	n.blockSize = size
}

func (n *Node) SetNumConnections(num int) {
	if num < 0 {
		num = 0
	}
	n.numConnections = num
}

func (n *Node) NumConnections() int {
	return n.numConnections
}

func (n *Node) OutboundCount() int {
	return len(n.outbound)
}

func (n *Node) SetChurnNode(churn bool) {
	n.churnNode = churn
}

func (n *Node) SetRNG(rng randomSource) {
	if rng != nil {
		n.rng = rng
	}
}

func (n *Node) SetBlockAcceptedObserver(observer func(self *Node, block *core.Block, timestamp core.SimTime)) {
	n.onBlockAccepted = observer
}

func (n *Node) SetFlowObserver(observer func(from, to *Node, block *core.Block, transmission, reception core.SimTime)) {
	n.onFlowBlock = observer
}

func (n *Node) AddNeighbor(peer *Node) bool {
	if peer == nil || peer == n {
		return false
	}
	for _, existing := range n.outbound {
		if existing == peer {
			return false
		}
	}
	for _, existing := range n.inbound {
		if existing == peer {
			return false
		}
	}
	if len(n.outbound) >= n.numConnections {
		return false
	}
	n.outbound = append(n.outbound, peer)
	return peer.addInbound(n)
}

func (n *Node) addInbound(peer *Node) bool {
	if peer == nil || peer == n {
		return false
	}
	n.inbound = append(n.inbound, peer)
	return true
}

func (n *Node) Neighbors() []*Node {
	out := make([]*Node, 0, len(n.outbound)+len(n.inbound))
	out = append(out, n.outbound...)
	out = append(out, n.inbound...)
	return out
}

func (n *Node) JoinNetwork(nodes []*Node, rng randomSource) {
	if len(nodes) == 0 || n.numConnections <= 0 {
		return
	}
	candidates := make([]int, 0, len(nodes))
	for i := range nodes {
		candidates = append(candidates, i)
	}
	if rng != nil {
		rng.Shuffle(len(candidates), func(i, j int) {
			candidates[i], candidates[j] = candidates[j], candidates[i]
		})
	}
	for _, idx := range candidates {
		if len(n.outbound) >= n.numConnections {
			break
		}
		_ = n.AddNeighbor(nodes[idx])
	}
}

func (n *Node) ID() int {
	return n.id
}

func (n *Node) Region() int {
	return n.region
}

func (n *Node) HashPower() uint64 {
	return n.hashPower
}

func (n *Node) BlockSize() uint64 {
	return n.blockSize
}

func (n *Node) Tip() *core.Block {
	return n.tip
}

func (n *Node) Orphans() []*core.Block {
	out := make([]*core.Block, 0, len(n.orphans))
	for _, b := range n.orphans {
		out = append(out, b)
	}
	return out
}

func (n *Node) SupportsCompactBlockRelay() bool {
	return n.useCompactBlockRelay
}

func (n *Node) ReceiveBlock(b *core.Block) bool {
	if b == nil {
		return false
	}

	if n.consensus == nil {
		if n.tip == nil {
			n.tip = b
			n.recordBlockAccepted(b)
			n.SendInv(b)
			return true
		}
		if b.Height() > n.tip.Height() {
			n.tip = b
			n.recordBlockAccepted(b)
			n.SendInv(b)
			return true
		}
		if b.Height() == n.tip.Height() {
			if b.ID() < n.tip.ID() {
				n.tip = b
				n.recordBlockAccepted(b)
				n.SendInv(b)
				return true
			}
			n.orphans[b.ID()] = b
			return false
		}
		n.orphans[b.ID()] = b
		return false
	}

	if n.consensus != nil {
		if !n.isBlockValidAgainstForkChoice(b) {
			if _, seen := n.orphans[b.ID()]; !seen && (n.tip == nil || !b.IsOnSameChainAs(n.tip)) {
				n.addOrphans(b, n.tip)
			}
			return false
		}
		if !n.isValidUncleSet(b) {
			return false
		}
		n.rememberKnownChain(b)
	}

	if n.forkChoice == ForkChoiceGHOST {
		return n.receiveBlockByGHOST()
	}

	if n.tip != nil && !n.tip.IsOnSameChainAs(b) {
		n.addOrphans(n.tip, b)
	}

	n.tip = b
	n.recordBlockAccepted(b)
	n.rebuildIncludedUncles()
	n.startMinting()
	n.SendInv(b)
	return true
}

func (n *Node) isBlockValidAgainstForkChoice(b *core.Block) bool {
	if n.consensus == nil {
		return false
	}
	switch n.forkChoice {
	case ForkChoiceGHOST:
		return n.consensus.IsReceivedBlockValid(b, nil)
	default:
		return n.consensus.IsReceivedBlockValid(b, n.tip)
	}
}

func (n *Node) receiveBlockByGHOST() bool {
	nextTip := n.ghostTip()
	if nextTip == nil {
		return false
	}
	if n.tip != nil && n.tip.ID() == nextTip.ID() {
		return false
	}
	if n.tip != nil && !n.tip.IsOnSameChainAs(nextTip) {
		n.addOrphans(n.tip, nextTip)
	}
	n.tip = nextTip
	n.recordBlockAccepted(nextTip)
	n.rebuildIncludedUncles()
	n.startMinting()
	n.SendInv(nextTip)
	return true
}

func (n *Node) rememberKnownChain(b *core.Block) {
	for cur := b; cur != nil; cur = cur.Parent() {
		n.knownBlocks[cur.ID()] = cur
	}
	for _, u := range b.Uncles() {
		for cur := u; cur != nil; cur = cur.Parent() {
			n.knownBlocks[cur.ID()] = cur
		}
	}
}

func (n *Node) ghostTip() *core.Block {
	if len(n.knownBlocks) == 0 {
		return nil
	}
	children := make(map[uint64][]uint64, len(n.knownBlocks))
	roots := make([]uint64, 0, 1)
	for id, b := range n.knownBlocks {
		parentID, hasParent := b.ParentID()
		if !hasParent {
			roots = append(roots, id)
			continue
		}
		if _, exists := n.knownBlocks[parentID]; !exists {
			roots = append(roots, id)
			continue
		}
		children[parentID] = append(children[parentID], id)
	}

	weightMemo := make(map[uint64]uint64, len(n.knownBlocks))
	var subtreeWeight func(id uint64) uint64
	subtreeWeight = func(id uint64) uint64 {
		if w, ok := weightMemo[id]; ok {
			return w
		}
		total := nodeWeight(n.knownBlocks[id])
		for _, childID := range children[id] {
			total += subtreeWeight(childID)
		}
		weightMemo[id] = total
		return total
	}

	best := roots[0]
	for _, id := range roots[1:] {
		best = n.pickHeavierSubtree(best, id, subtreeWeight)
	}

	for {
		kids := children[best]
		if len(kids) == 0 {
			break
		}
		bestChild := kids[0]
		for _, candidate := range kids[1:] {
			bestChild = n.pickHeavierSubtree(bestChild, candidate, subtreeWeight)
		}
		best = bestChild
	}

	return n.knownBlocks[best]
}

func (n *Node) pickHeavierSubtree(left, right uint64, subtreeWeight func(id uint64) uint64) uint64 {
	lw := subtreeWeight(left)
	rw := subtreeWeight(right)
	if rw > lw {
		return right
	}
	if lw > rw {
		return left
	}
	lWork := blockWork(n.knownBlocks[left])
	rWork := blockWork(n.knownBlocks[right])
	if rWork > lWork {
		return right
	}
	if lWork > rWork {
		return left
	}
	if right < left {
		return right
	}
	return left
}

func blockWork(b *core.Block) uint64 {
	if data, ok := consensus.PoWDataFromBlock(b); ok {
		return data.TotalDifficulty
	}
	return b.Height()
}

func canonicalChainSet(tip *core.Block) map[uint64]struct{} {
	out := make(map[uint64]struct{})
	for cur := tip; cur != nil; cur = cur.Parent() {
		out[cur.ID()] = struct{}{}
	}
	return out
}

func blockDifficulty(b *core.Block) uint64 {
	if data, ok := consensus.PoWDataFromBlock(b); ok {
		return data.Difficulty
	}
	if b == nil {
		return 0
	}
	if b.Height() == 0 {
		return 0
	}
	return 1
}

func nodeWeight(b *core.Block) uint64 {
	if b == nil {
		return 0
	}
	weight := blockDifficulty(b)
	for _, u := range b.Uncles() {
		weight += blockDifficulty(u)
	}
	return weight
}

func (n *Node) isValidUncleSet(b *core.Block) bool {
	if n.forkChoice != ForkChoiceGHOST {
		return true
	}
	uncles := b.Uncles()
	if len(uncles) > maxUnclesPerBlock {
		return false
	}
	if len(uncles) == 0 {
		return true
	}
	parent := b.Parent()
	if parent == nil {
		return false
	}
	parentChain := canonicalChainSet(parent)
	seen := make(map[uint64]struct{}, len(uncles))
	for _, u := range uncles {
		if u == nil {
			return false
		}
		if u.ID() == parent.ID() {
			return false
		}
		if _, dup := seen[u.ID()]; dup {
			return false
		}
		seen[u.ID()] = struct{}{}
		if _, already := n.includedUncles[u.ID()]; already {
			return false
		}
		if _, onParentChain := parentChain[u.ID()]; onParentChain {
			return false
		}
		parentID, hasParent := u.ParentID()
		if !hasParent {
			return false
		}
		if _, parentOnChain := parentChain[parentID]; !parentOnChain {
			return false
		}
		if b.Height() <= u.Height() {
			return false
		}
		if b.Height()-u.Height() > maxUncleGenerations {
			return false
		}
	}
	return true
}

func (n *Node) rebuildIncludedUncles() {
	next := make(map[uint64]struct{})
	for cur := n.tip; cur != nil; cur = cur.Parent() {
		for _, u := range cur.Uncles() {
			if u == nil {
				continue
			}
			next[u.ID()] = struct{}{}
		}
	}
	n.includedUncles = next
}

func (n *Node) selectUnclesForChild(parent *core.Block) []*core.Block {
	if n.forkChoice != ForkChoiceGHOST || parent == nil {
		return nil
	}
	parentChain := canonicalChainSet(parent)
	candidates := make([]*core.Block, 0, len(n.knownBlocks))
	newHeight := parent.Height() + 1
	for _, b := range n.knownBlocks {
		if b == nil {
			continue
		}
		if b.ID() == parent.ID() {
			continue
		}
		if _, onParentChain := parentChain[b.ID()]; onParentChain {
			continue
		}
		if _, already := n.includedUncles[b.ID()]; already {
			continue
		}
		parentID, hasParent := b.ParentID()
		if !hasParent {
			continue
		}
		if _, parentOnChain := parentChain[parentID]; !parentOnChain {
			continue
		}
		if newHeight <= b.Height() {
			continue
		}
		if newHeight-b.Height() > maxUncleGenerations {
			continue
		}
		candidates = append(candidates, b)
	}
	sort.Slice(candidates, func(i, j int) bool {
		wi := blockDifficulty(candidates[i])
		wj := blockDifficulty(candidates[j])
		if wi != wj {
			return wi > wj
		}
		if candidates[i].Height() != candidates[j].Height() {
			return candidates[i].Height() > candidates[j].Height()
		}
		return candidates[i].ID() < candidates[j].ID()
	})
	if len(candidates) > maxUnclesPerBlock {
		candidates = candidates[:maxUnclesPerBlock]
	}
	return candidates
}

func (n *Node) addOrphans(orphanBlock, validBlock *core.Block) {
	if orphanBlock == nil || orphanBlock == validBlock {
		return
	}
	n.orphans[orphanBlock.ID()] = orphanBlock
	if validBlock != nil {
		delete(n.orphans, validBlock.ID())
	}
	if validBlock == nil || orphanBlock.Height() > validBlock.Height() {
		n.addOrphans(orphanBlock.Parent(), validBlock)
		return
	}
	if orphanBlock.Height() == validBlock.Height() {
		n.addOrphans(orphanBlock.Parent(), validBlock.Parent())
		return
	}
	n.addOrphans(orphanBlock, validBlock.Parent())
}

func (n *Node) startMinting() {
	if n.consensus == nil || n.timer == nil || n.tip == nil {
		return
	}
	intervalTask := n.consensus.Minting(n.tip, n.id, n.hashPower)
	if intervalTask == nil {
		return
	}

	if n.mintingTask != nil {
		n.timer.RemoveTask(n.mintingTask)
	}

	mineTask := tasks.NewMiningTaskWithParent(intervalTask.Interval(), n.tip.Height(), func() {
		var block *core.Block
		if builder, ok := n.consensus.(blockBuilderWithUncles); ok {
			block = builder.BuildChildBlockWithUncles(n.tip, n.id, n.timer.CurrentTime(), n.selectUnclesForChild(n.tip))
		} else {
			builder, ok := n.consensus.(blockBuilder)
			if !ok {
				return
			}
			block = builder.BuildChildBlock(n.tip, n.id, n.timer.CurrentTime())
		}
		if block != nil {
			n.ReceiveBlock(block)
		}
	})
	n.mintingTask = mineTask
	n.timer.PutTask(mineTask)
}

func (n *Node) SendInv(block *core.Block) {
	if n.timer == nil || n.network == nil || block == nil {
		return
	}
	for _, to := range n.Neighbors() {
		n.timer.PutTask(tasks.NewInvMessageTask(n, to, block, n.network))
	}
}

func (n *Node) RecordFlowBlock(to tasks.Endpoint, block *core.Block, interval core.SimTime) {
	if n.onFlowBlock == nil || n.timer == nil || block == nil {
		return
	}
	peer, ok := to.(*Node)
	if !ok {
		return
	}
	reception := n.timer.CurrentTime()
	transmission := reception - interval
	n.onFlowBlock(n, peer, block, transmission, reception)
}

func (n *Node) ReceiveMessage(message tasks.MessageTask) {
	if message == nil {
		return
	}

	from := message.From()
	block := message.Block()

	switch m := message.(type) {
	case *tasks.InvMessageTask:
		if block == nil {
			return
		}
		if n.tip != nil && block.ID() == n.tip.ID() {
			return
		}
		if _, known := n.downloadingBlocks[block.ID()]; known {
			return
		}
		if _, orphan := n.orphans[block.ID()]; orphan {
			return
		}
		if n.tip != nil {
			if n.consensus != nil {
				if !n.isBlockValidAgainstForkChoice(block) && block.IsOnSameChainAs(n.tip) {
					return
				}
			} else {
				if block.Height() <= n.tip.Height() && block.IsOnSameChainAs(n.tip) {
					return
				}
			}
		}
		peer, ok := from.(*Node)
		if !ok {
			return
		}
		n.downloadingBlocks[block.ID()] = block
		n.timer.PutTask(tasks.NewRecMessageTask(n, peer, block, n.network))
	case *tasks.RecMessageTask:
		n.messageQueue = append(n.messageQueue, m)
		if !n.sendingBlock {
			n.SendNextBlockMessage()
		}
	case *tasks.GetBlockTxnMessageTask:
		n.messageQueue = append(n.messageQueue, m)
		if !n.sendingBlock {
			n.SendNextBlockMessage()
		}
	case *tasks.CmpctBlockMessageTask:
		if block == nil {
			return
		}
		if n.rng.Float64() > n.cbrFailureRate {
			delete(n.downloadingBlocks, block.ID())
			n.ReceiveBlock(block)
			return
		}
		peer, ok := from.(*Node)
		if !ok {
			return
		}
		n.timer.PutTask(tasks.NewGetBlockTxnMessageTask(n, peer, block, n.network))
	case *tasks.BlockMessageTask:
		if block == nil {
			return
		}
		delete(n.downloadingBlocks, block.ID())
		n.ReceiveBlock(block)
	}
}

func (n *Node) SendNextBlockMessage() {
	if n.timer == nil || n.network == nil {
		return
	}
	if len(n.messageQueue) == 0 {
		n.sendingBlock = false
		return
	}

	msg := n.messageQueue[0]
	n.messageQueue = n.messageQueue[1:]

	to, ok := msg.From().(*Node)
	if !ok {
		n.sendingBlock = false
		return
	}
	block := msg.Block()
	if block == nil {
		n.sendingBlock = false
		return
	}

	var task core.Task
	switch msg.(type) {
	case *tasks.RecMessageTask:
		if to.SupportsCompactBlockRelay() && n.useCompactBlockRelay {
			delay := n.network.TransferTime(n.compactSize, n.region, to.region) + n.processingTime
			task = tasks.NewCmpctBlockMessageTask(n, to, block, delay, n.network)
		} else {
			delay := n.network.TransferTime(n.blockSize, n.region, to.region) + n.processingTime
			task = tasks.NewBlockMessageTask(n, to, block, delay, n.network)
		}
	case *tasks.GetBlockTxnMessageTask:
		delay := n.network.TransferTime(n.failedBlockSize(), n.region, to.region) + n.processingTime
		task = tasks.NewBlockMessageTask(n, to, block, delay, n.network)
	default:
		n.sendingBlock = false
		return
	}

	n.sendingBlock = true
	n.timer.PutTask(task)
}

func (n *Node) failedBlockSize() uint64 {
	var ratio float64
	if n.churnNode {
		ratio = sampleChurnFailureRatio(n.rng.Intn(945))
	} else {
		// Java still consumes one random integer even though all entries are 0.01.
		_ = n.rng.Intn(210)
		ratio = 0.01
	}
	size := uint64(float64(n.blockSize) * ratio)
	if size == 0 {
		size = 1
	}
	return size
}

func sampleChurnFailureRatio(index int) float64 {
	acc := 0
	for _, bucket := range churnFailureBuckets {
		acc += bucket.count
		if index < acc {
			return bucket.ratio
		}
	}
	return churnFailureBuckets[len(churnFailureBuckets)-1].ratio
}

func (n *Node) recordBlockAccepted(block *core.Block) {
	if n.onBlockAccepted == nil || n.timer == nil || block == nil {
		return
	}
	n.onBlockAccepted(n, block, n.timer.CurrentTime())
}
