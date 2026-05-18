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
  fork_choice: heaviest
  end_time: 100000000
  end_block_height: 3
  java_compatible: true

network:
  profile: bitcoin_2019
  latency_matrix_file: ./data/latency.txt
  upload_bandwidth: [1, 2]
  download_bandwidth: [3, 4]
  region_distribution: [0.5, 0.5]
  degree_distribution: [0.5, 1.0]
`)
	if err := os.WriteFile(configPath, content, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := LoadConfig([]string{
		"--config", configPath,
		"--num-nodes", "12",
		"--block-interval", "1234",
		"--block-size", "4321",
		"--fork-choice", "ghost",
		"--end-time", "9999",
		"--end-block-height", "7",
		"--java-compatible", "false",
		"--latency-matrix-file", "./custom/latency.txt",
		"--network-profile", "custom_profile",
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
	if cfg.Simulation.ForkChoice != "ghost" {
		t.Fatalf("fork choice = %q, want %q", cfg.Simulation.ForkChoice, "ghost")
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
	if cfg.Network.Profile != "custom_profile" {
		t.Fatalf("network profile = %q, want %q", cfg.Network.Profile, "custom_profile")
	}
	if len(cfg.Network.UploadBandwidth) != 2 || cfg.Network.UploadBandwidth[0] != 1 || cfg.Network.UploadBandwidth[1] != 2 {
		t.Fatalf("upload bandwidth = %v, want [1 2]", cfg.Network.UploadBandwidth)
	}
	if len(cfg.Network.DownloadBandwidth) != 2 || cfg.Network.DownloadBandwidth[0] != 3 || cfg.Network.DownloadBandwidth[1] != 4 {
		t.Fatalf("download bandwidth = %v, want [3 4]", cfg.Network.DownloadBandwidth)
	}
	if len(cfg.Network.RegionDistribution) != 2 || cfg.Network.RegionDistribution[0] != 0.5 || cfg.Network.RegionDistribution[1] != 0.5 {
		t.Fatalf("region distribution = %v, want [0.5 0.5]", cfg.Network.RegionDistribution)
	}
	if len(cfg.Network.DegreeDistribution) != 2 || cfg.Network.DegreeDistribution[0] != 0.5 || cfg.Network.DegreeDistribution[1] != 1.0 {
		t.Fatalf("degree distribution = %v, want [0.5 1]", cfg.Network.DegreeDistribution)
	}
}
