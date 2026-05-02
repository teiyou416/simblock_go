package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Simulation struct {
		NumNodes      int   `mapstructure:"num_nodes"`
		BlockInterval int64 `mapstructure:"block_interval"`
		BlockSize     int   `mapstructure:"block_size"`
		EndTime       int64 `mapstructure:"end_time"`
	} `mapstructure:"simulation"`
	Network struct {
		LatencyMatrixFile string `mapstructure:"latency_matrix_file"`
	} `mapstructure:"network"`
}

var GlobalConfig Config

// InitConfig 负责在程序启动时读取 YAML
func InitConfig() {
	viper.SetConfigName("simulator")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config") // 查找配置文件的路径
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	if err := viper.Unmarshal(&GlobalConfig); err != nil {
		log.Fatalf("Unable to decode into struct, %v", err)
	}

	log.Printf("Configuration loaded. Total nodes: %d", GlobalConfig.Simulation.NumNodes)
}
