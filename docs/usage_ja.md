# SimBlock-Go 使用方法ガイド（日本語）

## 1. 前提条件

- Go 1.22 以上（または `go.mod` で指定されたバージョン）
- `make`

## 2. ビルドと実行

```bash
make build
make run
```

直接実行することもできます。

```bash
go run ./cmd/simblock
```

パラメータは [config/simulator.yaml](../config/simulator.yaml) の同名設定を上書きします。例：

```bash
go run ./cmd/simblock --num-nodes 100 --block-interval 300000 --java-compatible false
```

別の YAML 設定ファイルを読み込む場合は、`--config` を指定します。

```bash
go run ./cmd/simblock --config ./config/simulator.yaml --latency-matrix-file ./data/latency.txt
```

デフォルトの出力先は `output/` です。

## 3. パラメータ

シミュレータ本体は、`config/simulator.yaml` のパラメータ値をコマンドライン引数で上書きできます。

対応している引数：

- `--config`: YAML 設定ファイルのパス
- `--num-nodes`: `simulation.num_nodes` を上書き
- `--block-interval`: `simulation.block_interval` を上書き
- `--block-size`: `simulation.block_size` を上書き
- `--end-time`: `simulation.end_time` を上書き
- `--end-block-height`: `simulation.end_block_height` を上書き
- `--java-compatible`: `simulation.java_compatible` を上書き
- `--network-profile`: `network.profile` を上書き
- `--latency-matrix-file`: `network.latency_matrix_file` を上書き

## 4. シミュレーション設定

メイン設定ファイル：

- `config/simulator.yaml`

主な設定項目：

- `simulation.num_nodes`: シミュレーションするノード数
- `simulation.block_interval`: 平均ブロック生成間隔（ミリ秒）
- `simulation.block_size`: ブロックサイズ（バイト）
- `simulation.end_time`: 通常の Go モードで使う終了時刻
- `simulation.end_block_height`: Java 互換モードで使う終了ブロック高
- `simulation.java_compatible`: Java SimBlock 互換挙動を有効にするか
- `network.profile`: 組み込みネットワーク profile 名。現在は `bitcoin_2019` をサポート
- `network.latency_matrix_file`: ネットワーク遅延行列ファイルのパス
- `network.upload_bandwidth`: 各リージョンのアップロード帯域（bit/s）
- `network.download_bandwidth`: 各リージョンのダウンロード帯域（bit/s）
- `network.region_distribution`: リージョンごとのノード分布。合計は `1.0` を推奨
- `network.degree_distribution`: outbound link 数の累積分布。最後の値は `1.0`

通常の Go シミュレーションでは `java_compatible: false` を推奨します。Java 版 SimBlock に近い挙動が必要な場合だけ `java_compatible: true` を使います。

## 5. テスト

ユニットテストと統合テストを実行します。

```bash
make test
```

補助スクリプトからも実行できます。

```bash
./scripts/run_tests.sh
```

## 6. 出力ファイル

主な出力ファイル：

- `output/output.json`
- `output/static.json`
- `output/metrics.json`

これらのファイルは実行時に生成され、通常は Git 管理対象に含めません。
