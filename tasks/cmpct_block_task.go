package tasks

import (
	"github.com/teiyou416/simblock_go/core"
	"github.com/teiyou416/simblock_go/network"
)

type CmpctBlockMessageTask struct {
	baseMessageTask
}

func NewCmpctBlockMessageTask(
	from, to Endpoint,
	block *core.Block,
	transferDelay core.SimTime,
	net *network.Model,
) *CmpctBlockMessageTask {
	interval := transferDelay
	if net != nil {
		interval += net.Latency(from.Region(), to.Region())
	}
	return &CmpctBlockMessageTask{
		baseMessageTask: baseMessageTask{
			from:     from,
			to:       to,
			block:    block,
			interval: interval,
		},
	}
}

func (t *CmpctBlockMessageTask) Run() {
	t.from.SendNextBlockMessage()
	t.from.RecordFlowBlock(t.to, t.block, t.interval)
	t.to.ReceiveMessage(t)
}
