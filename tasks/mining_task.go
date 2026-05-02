package tasks

import "github.com/teiyou416/simblock_go/core"

type MiningTask struct {
	interval core.SimTime
	onMine   func()
}

func NewMiningTask(interval core.SimTime, onMine func()) *MiningTask {
	return &MiningTask{
		interval: interval,
		onMine:   onMine,
	}
}

func (t *MiningTask) Interval() core.SimTime {
	return t.interval
}

func (t *MiningTask) Run() {
	if t.onMine != nil {
		t.onMine()
	}
}
