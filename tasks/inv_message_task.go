package tasks

import (
	"github.com/teiyou416/simblock_go/core"
	"github.com/teiyou416/simblock_go/network"
)

type InvMessageTask struct {
	baseMessageTask
}

func NewInvMessageTask(from, to Endpoint, block *core.Block, net *network.Model) *InvMessageTask {
	return &InvMessageTask{
		baseMessageTask: baseMessageTask{
			from:     from,
			to:       to,
			block:    block,
			interval: baseInterval(from, to, net),
		},
	}
}

func (t *InvMessageTask) Run() {
	t.to.ReceiveMessage(t)
}
