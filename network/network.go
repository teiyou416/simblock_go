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
	return m.GetBandwidth(fromRegion, toRegion)
}

// GetBandwidth returns the bottleneck bandwidth between two regions.
//
// Bandwidth unit: bit/second.
func (m *Model) GetBandwidth(fromRegion, toRegion int) uint64 {
	up := m.upload[fromRegion]
	down := m.download[toRegion]
	if up < down {
		return up
	}
	return down
}

// TransferTime returns payload transfer time (without propagation latency).
//
// sizeBytes unit: byte
// return unit: milliseconds (simulation time)
func (m *Model) TransferTime(sizeBytes uint64, fromRegion, toRegion int) core.SimTime {
	if sizeBytes == 0 {
		return 0
	}
	bandwidth := m.GetBandwidth(fromRegion, toRegion)
	if bandwidth < 1000 {
		// Keep deterministic behavior and avoid division-by-zero.
		return 0
	}
	return core.SimTime((sizeBytes * 8) / (bandwidth / 1000))
}
