package consensus

import "github.com/teiyou416/simblock_go/core"

// Algorithm defines consensus-specific behavior expected by the node layer.
//
// Concrete implementations (e.g., PoW) are responsible for translating local
// chain state into minting events and validating incoming blocks.
type Algorithm interface {
	// Minting creates a minting task for the current tip.
	//
	// Returned task should be scheduled by the simulator timer.
	Minting(currentTip *core.Block, selfNodeID int, selfHashPower uint64) core.Task

	// IsReceivedBlockValid decides whether the received block should be accepted
	// as a canonical-tip candidate against currentTip.
	IsReceivedBlockValid(receivedBlock, currentTip *core.Block) bool

	// GenesisBlock returns the chain genesis block for this algorithm.
	GenesisBlock(minterID int) *core.Block
}
