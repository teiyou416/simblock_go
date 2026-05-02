package app

import (
	"fmt"
	"log"

	"github.com/teiyou416/simblock_go/config"
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

	// TODO: initialize global engine (timer/scheduler)
	// TODO: initialize network topology
	// TODO: start simulator main loop
}
