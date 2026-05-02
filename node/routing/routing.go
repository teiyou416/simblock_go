package routing

// NodeRef is the minimal node view required by routing tables.
//
// It avoids importing the concrete node package here, keeping routing
// abstractions decoupled from node implementations.
type NodeRef interface {
	ID() int
}

// Table defines routing behavior used by a node.
//
// This mirrors the Java AbstractRoutingTable contract while keeping Go
// composition-friendly.
type Table interface {
	InitTable()
	Neighbors() []NodeRef
	AddNeighbor(node NodeRef) bool
	RemoveNeighbor(node NodeRef) bool

	SetNumConnection(numConnection int)
	NumConnection() int

	AddInbound(from NodeRef) bool
	RemoveInbound(from NodeRef) bool
	AcceptBlock()
}

// BaseTable provides common/default behavior for routing tables.
//
// Concrete implementations can embed this struct and override methods as
// needed.
type BaseTable struct {
	self          NodeRef
	numConnection int
}

func NewBaseTable(self NodeRef) *BaseTable {
	return &BaseTable{
		self:          self,
		numConnection: 8,
	}
}

func (t *BaseTable) Self() NodeRef {
	return t.self
}

func (t *BaseTable) SetNumConnection(numConnection int) {
	if numConnection < 0 {
		numConnection = 0
	}
	t.numConnection = numConnection
}

func (t *BaseTable) NumConnection() int {
	return t.numConnection
}

func (t *BaseTable) AddInbound(from NodeRef) bool {
	return false
}

func (t *BaseTable) RemoveInbound(from NodeRef) bool {
	return false
}

func (t *BaseTable) AcceptBlock() {}
