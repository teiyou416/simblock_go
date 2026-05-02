package tasks

import (
	"github.com/teiyou416/simblock_go/core"
	"github.com/teiyou416/simblock_go/network"
)

type Endpoint interface {
	ID() int
	Region() int
	ReceiveMessage(MessageTask)
	SupportsCompactBlockRelay() bool
	SendNextBlockMessage()
	RecordFlowBlock(to Endpoint, block *core.Block, interval core.SimTime)
}

type MessageTask interface {
	core.Task
	From() Endpoint
	To() Endpoint
	Block() *core.Block
}

const baseMessageOverhead = core.SimTime(10)

type baseMessageTask struct {
	from     Endpoint
	to       Endpoint
	block    *core.Block
	interval core.SimTime
}

func (t *baseMessageTask) From() Endpoint     { return t.from }
func (t *baseMessageTask) To() Endpoint       { return t.to }
func (t *baseMessageTask) Block() *core.Block { return t.block }
func (t *baseMessageTask) Interval() core.SimTime {
	return t.interval
}

func baseInterval(from, to Endpoint, net *network.Model) core.SimTime {
	if net == nil {
		return baseMessageOverhead
	}
	return net.Latency(from.Region(), to.Region()) + baseMessageOverhead
}
