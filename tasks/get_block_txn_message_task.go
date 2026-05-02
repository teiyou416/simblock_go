package tasks

import (
	"github.com/teiyou416/simblock_go/core"
	"github.com/teiyou416/simblock_go/network"
)

type GetBlockTxnMessageTask struct {
	baseMessageTask
}

func NewGetBlockTxnMessageTask(from, to Endpoint, block *core.Block, net *network.Model) *GetBlockTxnMessageTask {
	return &GetBlockTxnMessageTask{
		baseMessageTask: baseMessageTask{
			from:     from,
			to:       to,
			block:    block,
			interval: baseInterval(from, to, net),
		},
	}
}

func (t *GetBlockTxnMessageTask) Run() {
	t.to.ReceiveMessage(t)
}
