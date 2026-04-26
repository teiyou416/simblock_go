package main

import (
	"fmt"
	"github.com/teiyou416/simblock_go/config"
)

func main() {
	fmt.Println("Starting SimBlock-Go...")

	// 第一步：加载配置
	config.InitConfig()

	// TODO: 初始化全局引擎 (Engine/Timer)
	// TODO: 初始化网络拓扑 (Network)
	// TODO: 启动模拟主循环 (Simulator Run)

}
