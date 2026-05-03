package tasks

import "github.com/teiyou416/simblock_go/core"

type MiningTask struct {
	interval     core.SimTime
	onMine       func()
	parentHeight uint64
	hasParent    bool
}

func NewMiningTask(interval core.SimTime, onMine func()) *MiningTask {
	return &MiningTask{
		interval: interval,
		onMine:   onMine,
	}
}

func NewMiningTaskWithParent(interval core.SimTime, parentHeight uint64, onMine func()) *MiningTask {
	return &MiningTask{
		interval:     interval,
		onMine:       onMine,
		parentHeight: parentHeight,
		hasParent:    true,
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

func (t *MiningTask) ParentHeight() (uint64, bool) {
	return t.parentHeight, t.hasParent
}
