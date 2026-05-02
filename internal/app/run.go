package app

import (
	"fmt"
	"log"

	"github.com/teiyou416/simblock_go/config"
	"github.com/teiyou416/simblock_go/core"
	"github.com/teiyou416/simblock_go/engine"
	"github.com/teiyou416/simblock_go/network"
)

// Run boots the current SimBlock-Go application skeleton.
func Run() {
	fmt.Println("Starting SimBlock-Go...")

	// Step 1: load configuration.
	config.InitConfig()

	// Step 2: load network latency matrix from configuration.
	latency, err := network.LoadLatencyMatrix(config.GlobalConfig.Network.LatencyMatrixFile)
	if err != nil {
		log.Fatalf("failed to load latency matrix: %v", err)
	}
	log.Printf("Latency matrix loaded: %dx%d", len(latency), len(latency[0]))

	upload, download := defaultBandwidths(len(latency))
	netModel := network.NewModel(latency, upload, download)
	timer := engine.NewTimer()

	sim := engine.NewSimulator(engine.SimulatorConfig{
		NumNodes:           config.GlobalConfig.Simulation.NumNodes,
		TargetInterval:     core.SimTime(config.GlobalConfig.Simulation.BlockInterval),
		EndTime:            core.SimTime(config.GlobalConfig.Simulation.EndTime),
		BlockSize:          uint64(config.GlobalConfig.Simulation.BlockSize),
		OutputDir:          "output",
		RandomSeed:         10,
		ConnectionsPerNode: 8,
	}, timer, netModel)

	stats, err := sim.Run()
	if err != nil {
		log.Fatalf("simulation run failed: %v", err)
	}
	log.Printf(
		"Simulation finished. accepted_blocks=%d observed_blocks=%d mean_delay=%.2f orphan_rate=%.4f",
		stats.AcceptedBlocks,
		stats.ObservedBlocks,
		stats.MeanPropagationDelay,
		stats.OrphanRate,
	)
}

func defaultBandwidths(regionCount int) ([]uint64, []uint64) {
	uploadDefaults := []uint64{19_200_000, 20_700_000, 5_800_000, 15_700_000, 10_200_000, 11_300_000}
	downloadDefaults := []uint64{52_000_000, 40_000_000, 18_000_000, 22_800_000, 22_800_000, 29_900_000}

	upload := make([]uint64, regionCount)
	download := make([]uint64, regionCount)
	for i := 0; i < regionCount; i++ {
		upload[i] = uploadDefaults[i%len(uploadDefaults)]
		download[i] = downloadDefaults[i%len(downloadDefaults)]
	}
	return upload, download
}
