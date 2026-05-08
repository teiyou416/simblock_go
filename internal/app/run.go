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
func Run(args []string) {
	fmt.Println("Starting SimBlock-Go...")

	// Step 1: load configuration.
	config.InitConfig(args)

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
		EndBlockHeight:     config.GlobalConfig.Simulation.EndBlockHeight,
		BlockSize:          uint64(config.GlobalConfig.Simulation.BlockSize),
		OutputDir:          "output",
		RandomSeed:         10,
		ConnectionsPerNode: 8,
		JavaCompatible:     config.GlobalConfig.Simulation.JavaCompatible,
	}, timer, netModel)

	stats, err := sim.Run()
	if err != nil {
		log.Fatalf("simulation run failed: %v", err)
	}
	log.Printf("Simulation finished.")
	log.Printf(
		"  events: total=%d add-node=%d add-link=%d add-block=%d flow-block=%d simulation-end=%d",
		stats.TotalEvents,
		stats.AddNodeEvents,
		stats.AddLinkEvents,
		stats.AddBlockEvents,
		stats.FlowBlockEvents,
		stats.SimulationEndEvents,
	)
	log.Printf("  simulation_end_timestamp=%d", stats.SimulationEndTime)
	log.Printf(
		"  blocks: accepted=%d observed=%d orphan_blocks=%d orphan_rate=%.4f",
		stats.AcceptedBlocks,
		stats.ObservedBlocks,
		stats.OrphanBlocks,
		stats.OrphanRate,
	)
	log.Printf("  mean_propagation_delay=%.2f", stats.MeanPropagationDelay)
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
