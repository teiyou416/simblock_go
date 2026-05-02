package tasks

import (
	"github.com/teiyou416/simblock_go/core"
	"github.com/teiyou416/simblock_go/network"
)

type RecMessageTask struct {
	baseMessageTask
}

func NewRecMessageTask(from, to Endpoint, block *core.Block, net *network.Model) *RecMessageTask {
	return &RecMessageTask{
		baseMessageTask: baseMessageTask{
			from:     from,
			to:       to,
			block:    block,
			interval: baseInterval(from, to, net),
		},
	}
}

func (t *RecMessageTask) Run() {
	t.to.ReceiveMessage(t)
}
