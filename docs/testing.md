# SimBlock-Go 测试说明

这个仓库现在有两个层次的测试：

- `unit tests`：各包自己的单元测试（`block/`, `core/`, `engine/`, `network/`, `node/`, `tasks/`）
- `integrated suite`：统一场景测试（`tests/integration_suite_test.go`）

## 推荐执行方式

1. 只跑 Go 测试（推荐日常开发）

```bash
make test
```

2. 一次性跑完整测试脚本

```bash
./scripts/run_tests.sh
```

## 目标

- 单元测试用于快速定位某个模块问题
- 集成测试用于保证关键链路（调度器、PoW 挖矿、模拟输出）整体可用
