package block

import "github.com/teiyou416/simblock_go/core"

var nextBlockID uint64

// Block is the minimal chain unit for the Go port.
type Block struct {
	id       uint64
	height   uint64
	parent   *Block
	minterID int
	time     core.SimTime
}

func NewBlock(parent *Block, minterID int, time core.SimTime) *Block {
	height := uint64(0)
	if parent != nil {
		height = parent.height + 1
	}

	b := &Block{
		id:       nextBlockID,
		height:   height,
		parent:   parent,
		minterID: minterID,
		time:     time,
	}
	nextBlockID++
	return b
}

func Genesis(minterID int) *Block {
	return NewBlock(nil, minterID, 0)
}

func (b *Block) ID() uint64 {
	return b.id
}

func (b *Block) Height() uint64 {
	return b.height
}

func (b *Block) Parent() *Block {
	return b.parent
}

func (b *Block) MinterID() int {
	return b.minterID
}

func (b *Block) Time() core.SimTime {
	return b.time
}

func (b *Block) BlockWithHeight(height uint64) *Block {
	if b == nil {
		return nil
	}
	if b.height == height {
		return b
	}
	if b.parent == nil {
		return nil
	}
	return b.parent.BlockWithHeight(height)
}

func (b *Block) IsOnSameChainAs(other *Block) bool {
	if b == nil || other == nil {
		return false
	}

	if b.height <= other.height {
		sameHeight := other.BlockWithHeight(b.height)
		return sameHeight != nil && sameHeight.id == b.id
	}

	sameHeight := b.BlockWithHeight(other.height)
	return sameHeight != nil && sameHeight.id == other.id
}
