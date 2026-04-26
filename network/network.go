package network

import "github.com/teiyou416/simblock_go/core"

// Model is a deterministic network model for the simulator core.
type Model struct {
	latency  [][]core.SimTime
	upload   []uint64
	download []uint64
}

func NewModel(
	latency [][]core.SimTime,
	uploadBandwidth []uint64,
	downloadBandwidth []uint64,
) *Model {
	return &Model{
		latency:  latency,
		upload:   uploadBandwidth,
		download: downloadBandwidth,
	}
}

func (m *Model) Latency(fromRegion, toRegion int) core.SimTime {
	return m.latency[fromRegion][toRegion]
}

func (m *Model) Bandwidth(fromRegion, toRegion int) uint64 {
	up := m.upload[fromRegion]
	down := m.download[toRegion]
	if up < down {
		return up
	}
	return down
}
