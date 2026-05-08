# SimBlock-Go

[English](./README.md) | [中文](./README_zh.md) | [日本語](./README_ja.md)

SimBlock の Go 実装です。

## クイックスタート

```bash
make build
make run
```

`config/simulator.yaml` の設定は、コマンドライン引数で上書きできます。

```bash
go run ./cmd/simblock --num-nodes 100 --block-interval 300000 --java-compatible false
```

別の YAML 設定ファイルを使う場合は、`--config` を指定します。

```bash
go run ./cmd/simblock --config ./config/simulator.yaml --latency-matrix-file ./data/latency.txt
```

Java 版 SimBlock モードを使う場合は、次のように指定します。

```bash
go run ./cmd/simblock --num-nodes 100 --block-interval 300000 --java-compatible true --end-block-height 1000
```

## テスト

```bash
make test
```

## ドキュメント

詳しい使い方は以下を参照してください。

- [テストガイド](./docs/testing.md)
- [Usage Guide (English)](./docs/usage_en.md)
- [使用说明 (中文)](./docs/usage_zh.md)
- [使用方法ガイド (日本語)](./docs/usage_ja.md)
