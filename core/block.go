package core

var nextCoreBlockID uint64

// Block is the protocol-level chain unit.
//
// ConsensusData is intentionally untyped to allow different consensus
// implementations (PoW/PoS/...) to attach their own metadata.
type Block struct {
	id            uint64
	height        uint64
	parent        *Block
	uncles        []*Block
	minterID      int
	time          SimTime
	consensusData any
}

func NewBlock(parent *Block, minterID int, time SimTime, consensusData any) *Block {
	return NewBlockWithUncles(parent, minterID, time, consensusData, nil)
}

func NewBlockWithUncles(parent *Block, minterID int, time SimTime, consensusData any, uncles []*Block) *Block {
	height := uint64(0)
	if parent != nil {
		height = parent.height + 1
	}
	uncleCopy := make([]*Block, len(uncles))
	copy(uncleCopy, uncles)

	b := &Block{
		id:            nextCoreBlockID,
		height:        height,
		parent:        parent,
		uncles:        uncleCopy,
		minterID:      minterID,
		time:          time,
		consensusData: consensusData,
	}
	nextCoreBlockID++
	return b
}

func GenesisBlock(minterID int, consensusData any) *Block {
	return NewBlock(nil, minterID, 0, consensusData)
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

func (b *Block) Uncles() []*Block {
	if b == nil || len(b.uncles) == 0 {
		return nil
	}
	out := make([]*Block, len(b.uncles))
	copy(out, b.uncles)
	return out
}

func (b *Block) UncleIDs() []uint64 {
	if b == nil || len(b.uncles) == 0 {
		return nil
	}
	out := make([]uint64, 0, len(b.uncles))
	for _, u := range b.uncles {
		if u == nil {
			continue
		}
		out = append(out, u.id)
	}
	return out
}

func (b *Block) ParentID() (uint64, bool) {
	if b == nil || b.parent == nil {
		return 0, false
	}
	return b.parent.id, true
}

func (b *Block) MinterID() int {
	return b.minterID
}

func (b *Block) Time() SimTime {
	return b.time
}

func (b *Block) ConsensusData() any {
	if b == nil {
		return nil
	}
	return b.consensusData
}

// BlockWithHeight walks back via parent pointers until the target height.
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
