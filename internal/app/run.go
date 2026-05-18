package app

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/teiyou416/simblock_go/config"
	"github.com/teiyou416/simblock_go/core"
	"github.com/teiyou416/simblock_go/engine"
	"github.com/teiyou416/simblock_go/network"
)

// Run boots the current SimBlock-Go application skeleton.
func Run(args []string) {
	fmt.Println("Starting SimBlock-Go...")
	startedAt := time.Now()
	runOutputDir := filepath.Join("output", startedAt.Format("20060102_150405_000"))

	// Step 1: load configuration.
	config.InitConfig(args)

	// Step 2: load network latency matrix from configuration.
	latency, err := network.LoadLatencyMatrix(config.GlobalConfig.Network.LatencyMatrixFile)
	if err != nil {
		log.Fatalf("failed to load latency matrix: %v", err)
	}
	log.Printf("Latency matrix loaded: %dx%d", len(latency), len(latency[0]))

	profile, ok := network.ProfileByName(config.GlobalConfig.Network.Profile)
	if !ok {
		log.Fatalf("unknown network profile: %q", config.GlobalConfig.Network.Profile)
	}
	profile = profile.WithOverrides(network.ProfileOverrides{
		UploadBandwidth:    config.GlobalConfig.Network.UploadBandwidth,
		DownloadBandwidth:  config.GlobalConfig.Network.DownloadBandwidth,
		RegionDistribution: config.GlobalConfig.Network.RegionDistribution,
		DegreeDistribution: config.GlobalConfig.Network.DegreeDistribution,
	})
	if err := profile.Validate(len(latency)); err != nil {
		log.Fatalf("invalid network profile: %v", err)
	}
	upload, download := profile.Bandwidths(len(latency))
	netModel := network.NewModel(latency, upload, download)
	timer := engine.NewTimer()

	sim := engine.NewSimulator(engine.SimulatorConfig{
		NumNodes:           config.GlobalConfig.Simulation.NumNodes,
		TargetInterval:     core.SimTime(config.GlobalConfig.Simulation.BlockInterval),
		EndTime:            core.SimTime(config.GlobalConfig.Simulation.EndTime),
		EndBlockHeight:     config.GlobalConfig.Simulation.EndBlockHeight,
		BlockSize:          uint64(config.GlobalConfig.Simulation.BlockSize),
		OutputDir:          runOutputDir,
		RandomSeed:         10,
		ConnectionsPerNode: 8,
		JavaCompatible:     config.GlobalConfig.Simulation.JavaCompatible,
		NetworkProfile:     profile,
	}, timer, netModel)

	stats, err := sim.Run()
	if err != nil {
		log.Fatalf("simulation run failed: %v", err)
	}
	log.Printf("Simulation finished.")
	log.Printf("  wall_clock_duration=%s", time.Since(startedAt).Round(time.Millisecond))
	log.Printf("  output_dir=%s", runOutputDir)
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
