package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigAppliesCLIOverrides(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "simulator.yaml")
	content := []byte(`simulation:
  num_nodes: 300
  block_interval: 600000
  block_size: 535000
  end_time: 100000000
  end_block_height: 3
  java_compatible: true

network:
  latency_matrix_file: ./data/latency.txt
`)
	if err := os.WriteFile(configPath, content, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := LoadConfig([]string{
		"--config", configPath,
		"--num-nodes", "12",
		"--block-interval", "1234",
		"--block-size", "4321",
		"--end-time", "9999",
		"--end-block-height", "7",
		"--java-compatible", "false",
		"--latency-matrix-file", "./custom/latency.txt",
	})
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Simulation.NumNodes != 12 {
		t.Fatalf("num nodes = %d, want %d", cfg.Simulation.NumNodes, 12)
	}
	if cfg.Simulation.BlockInterval != 1234 {
		t.Fatalf("block interval = %d, want %d", cfg.Simulation.BlockInterval, 1234)
	}
	if cfg.Simulation.BlockSize != 4321 {
		t.Fatalf("block size = %d, want %d", cfg.Simulation.BlockSize, 4321)
	}
	if cfg.Simulation.EndTime != 9999 {
		t.Fatalf("end time = %d, want %d", cfg.Simulation.EndTime, 9999)
	}
	if cfg.Simulation.EndBlockHeight != 7 {
		t.Fatalf("end block height = %d, want %d", cfg.Simulation.EndBlockHeight, 7)
	}
	if cfg.Simulation.JavaCompatible {
		t.Fatalf("java compatible = %v, want false", cfg.Simulation.JavaCompatible)
	}
	if cfg.Network.LatencyMatrixFile != "./custom/latency.txt" {
		t.Fatalf("latency matrix file = %q, want %q", cfg.Network.LatencyMatrixFile, "./custom/latency.txt")
	}
}