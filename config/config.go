package config

import (
	"fmt"
	"io"
	"log"
	"strconv"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	Simulation struct {
		NumNodes       int    `mapstructure:"num_nodes"`
		BlockInterval  int64  `mapstructure:"block_interval"`
		BlockSize      int    `mapstructure:"block_size"`
		ForkChoice     string `mapstructure:"fork_choice"`
		EndTime        int64  `mapstructure:"end_time"`
		EndBlockHeight int    `mapstructure:"end_block_height"`
		JavaCompatible bool   `mapstructure:"java_compatible"`
		OutputMode     string `mapstructure:"output_mode"`
	} `mapstructure:"simulation"`
	Network struct {
		LatencyMatrixFile  string    `mapstructure:"latency_matrix_file"`
		Profile            string    `mapstructure:"profile"`
		UploadBandwidth    []uint64  `mapstructure:"upload_bandwidth"`
		DownloadBandwidth  []uint64  `mapstructure:"download_bandwidth"`
		RegionDistribution []float64 `mapstructure:"region_distribution"`
		DegreeDistribution []float64 `mapstructure:"degree_distribution"`
	} `mapstructure:"network"`
}

var GlobalConfig Config

// InitConfig 负责在程序启动时读取 YAML，并允许命令行参数覆盖同名配置。
func InitConfig(args []string) {
	cfg, err := LoadConfig(args)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	GlobalConfig = cfg
}

// LoadConfig loads the YAML file and applies explicit CLI overrides.
func LoadConfig(args []string) (Config, error) {
	flagSet := pflag.NewFlagSet("simblock", pflag.ContinueOnError)
	flagSet.SetOutput(io.Discard)

	configPath := flagSet.String("config", "./config/simulator.yaml", "path to the simulator YAML config")
	flagSet.Int("num-nodes", 0, "override simulation.num_nodes")
	flagSet.Int64("block-interval", 0, "override simulation.block_interval")
	flagSet.Int("block-size", 0, "override simulation.block_size")
	flagSet.String("fork-choice", "", "override simulation.fork_choice (heaviest|ghost)")
	flagSet.Int64("end-time", 0, "override simulation.end_time")
	flagSet.Int("end-block-height", 0, "override simulation.end_block_height")
	flagSet.String("java-compatible", "", "override simulation.java_compatible")
	flagSet.String("output-mode", "", "override simulation.output_mode (core|full)")
	flagSet.String("latency-matrix-file", "", "override network.latency_matrix_file")
	flagSet.String("network-profile", "", "override network.profile")

	if err := flagSet.Parse(args); err != nil {
		return Config{}, err
	}

	v := viper.New()
	v.SetConfigFile(*configPath)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return Config{}, fmt.Errorf("read config file %q: %w", *configPath, err)
	}

	if err := applyIntOverride(v, flagSet.Lookup("num-nodes"), "simulation.num_nodes"); err != nil {
		return Config{}, err
	}
	if err := applyInt64Override(v, flagSet.Lookup("block-interval"), "simulation.block_interval"); err != nil {
		return Config{}, err
	}
	if err := applyIntOverride(v, flagSet.Lookup("block-size"), "simulation.block_size"); err != nil {
		return Config{}, err
	}
	if err := applyStringOverride(v, flagSet.Lookup("fork-choice"), "simulation.fork_choice"); err != nil {
		return Config{}, err
	}
	if err := applyInt64Override(v, flagSet.Lookup("end-time"), "simulation.end_time"); err != nil {
		return Config{}, err
	}
	if err := applyIntOverride(v, flagSet.Lookup("end-block-height"), "simulation.end_block_height"); err != nil {
		return Config{}, err
	}
	if err := applyBoolOverride(v, flagSet.Lookup("java-compatible"), "simulation.java_compatible"); err != nil {
		return Config{}, err
	}
	if err := applyStringOverride(v, flagSet.Lookup("output-mode"), "simulation.output_mode"); err != nil {
		return Config{}, err
	}
	if err := applyStringOverride(v, flagSet.Lookup("latency-matrix-file"), "network.latency_matrix_file"); err != nil {
		return Config{}, err
	}
	if err := applyStringOverride(v, flagSet.Lookup("network-profile"), "network.profile"); err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("unable to decode config: %w", err)
	}

	return cfg, nil
}

func applyIntOverride(v *viper.Viper, flag *pflag.Flag, key string) error {
	if flag == nil || !flag.Changed {
		return nil
	}
	value, err := strconv.Atoi(flag.Value.String())
	if err != nil {
		return fmt.Errorf("parse %s: %w", flag.Name, err)
	}
	v.Set(key, value)
	return nil
}

func applyInt64Override(v *viper.Viper, flag *pflag.Flag, key string) error {
	if flag == nil || !flag.Changed {
		return nil
	}
	value, err := strconv.ParseInt(flag.Value.String(), 10, 64)
	if err != nil {
		return fmt.Errorf("parse %s: %w", flag.Name, err)
	}
	v.Set(key, value)
	return nil
}

func applyBoolOverride(v *viper.Viper, flag *pflag.Flag, key string) error {
	if flag == nil || !flag.Changed {
		return nil
	}
	value, err := strconv.ParseBool(flag.Value.String())
	if err != nil {
		return fmt.Errorf("parse %s: %w", flag.Name, err)
	}
	v.Set(key, value)
	return nil
}

func applyStringOverride(v *viper.Viper, flag *pflag.Flag, key string) error {
	if flag == nil || !flag.Changed {
		return nil
	}
	v.Set(key, flag.Value.String())
	return nil
}
