package node

import (
	"github.com/teiyou416/simblock_go/block"
	"github.com/teiyou416/simblock_go/core"
)

// Node is a minimal blockchain node model used by the simulator core.
type Node struct {
	id      int
	region  int
	tip     *block.Block
	orphans map[uint64]*block.Block
}

func New(id, region int) *Node {
	return &Node{
		id:      id,
		region:  region,
		orphans: make(map[uint64]*block.Block),
	}
}

func (n *Node) ID() int {
	return n.id
}

func (n *Node) Region() int {
	return n.region
}

func (n *Node) Tip() *block.Block {
	return n.tip
}

func (n *Node) Orphans() []*block.Block {
	out := make([]*block.Block, 0, len(n.orphans))
	for _, b := range n.orphans {
		out = append(out, b)
	}
	return out
}

func (n *Node) ReceiveBlock(b *block.Block) bool {
	if b == nil {
		return false
	}

	if n.tip == nil {
		n.tip = b
		return true
	}

	if b.Height() > n.tip.Height() {
		n.tip = b
		return true
	}

	if b.Height() == n.tip.Height() {
		// Deterministic tiebreaker for competing tips.
		if b.ID() < n.tip.ID() {
			n.tip = b
			return true
		}
		n.orphans[b.ID()] = b
		return false
	}

	n.orphans[b.ID()] = b
	return false
}

func (n *Node) MintBlock(now core.SimTime) *block.Block {
	newBlock := block.NewBlock(n.tip, n.id, now)
	n.tip = newBlock
	return newBlock
}
