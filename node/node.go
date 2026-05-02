package node

import (
	"math/rand"

	"github.com/teiyou416/simblock_go/core"
	"github.com/teiyou416/simblock_go/network"
	"github.com/teiyou416/simblock_go/node/consensus"
	"github.com/teiyou416/simblock_go/tasks"
)

const (
	defaultProcessingTime core.SimTime = 2
	defaultBlockSize      uint64       = 535000
	defaultCompactSize    uint64       = 18 * 1000
)

type scheduler interface {
	PutTask(task core.Task)
	PutTaskAt(task core.Task, timestamp core.SimTime)
	RemoveTask(task core.Task) bool
	CurrentTime() core.SimTime
}

type blockBuilder interface {
	BuildChildBlock(parent *core.Block, minterID int, now core.SimTime) *core.Block
}

// Node models a simulation node and protocol-side state transitions.
type Node struct {
	id        int
	region    int
	hashPower uint64

	tip     *core.Block
	orphans map[uint64]*core.Block

	neighbors []*Node

	timer   scheduler
	network *network.Model

	consensus consensus.Algorithm

	useCompactBlockRelay bool
	sendingBlock         bool
	messageQueue         []tasks.MessageTask
	downloadingBlocks    map[uint64]*core.Block
	mintingTask          core.Task

	processingTime core.SimTime
	blockSize      uint64
	compactSize    uint64
	cbrFailureRate float64
	rng            *rand.Rand

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
		rng:               rand.New(rand.NewSource(1)),
	}
}

func (n *Node) BindEnvironment(timer scheduler, net *network.Model) {
	n.timer = timer
	n.network = net
}

func (n *Node) SetConsensus(algo consensus.Algorithm) {
	n.consensus = algo
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
	for _, existing := range n.neighbors {
		if existing == peer {
			return false
		}
	}
	n.neighbors = append(n.neighbors, peer)
	return true
}

func (n *Node) Neighbors() []*Node {
	out := make([]*Node, len(n.neighbors))
	copy(out, n.neighbors)
	return out
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
		if !n.consensus.IsReceivedBlockValid(b, n.tip) {
			if n.tip == nil || !b.IsOnSameChainAs(n.tip) {
				n.orphans[b.ID()] = b
			}
			return false
		}
	}

	if n.tip != nil && !n.tip.IsOnSameChainAs(b) {
		n.orphans[n.tip.ID()] = n.tip
	}

	n.tip = b
	n.recordBlockAccepted(b)
	n.startMinting()
	n.SendInv(b)
	return true
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

	mineTask := tasks.NewMiningTask(intervalTask.Interval(), func() {
		builder, ok := n.consensus.(blockBuilder)
		if !ok {
			return
		}
		block := builder.BuildChildBlock(n.tip, n.id, n.timer.CurrentTime())
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
	for _, to := range n.neighbors {
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
				if !n.consensus.IsReceivedBlockValid(block, n.tip) && block.IsOnSameChainAs(n.tip) {
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
		delay := n.network.TransferTime(n.blockSize, n.region, to.region) + n.processingTime
		task = tasks.NewBlockMessageTask(n, to, block, delay, n.network)
	default:
		n.sendingBlock = false
		return
	}

	n.sendingBlock = true
	n.timer.PutTask(task)
}

func (n *Node) recordBlockAccepted(block *core.Block) {
	if n.onBlockAccepted == nil || n.timer == nil || block == nil {
		return
	}
	n.onBlockAccepted(n, block, n.timer.CurrentTime())
}
