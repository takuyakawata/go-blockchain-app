# Go Blockchain App

このプロジェクトは、Go言語で実装されたシンプルなブロックチェーンの学習用デモンストレーションです。3つの異なるパターンのブロックチェーン実装を提供しています。

## 概要

このプロジェクトでは以下の3つのブロックチェーンパターンを学習できます：

1. **Pattern 1**: シンプルなブロックチェーン（メモリベース）
2. **Pattern 2**: Proof of Work（PoW）機能付きブロックチェーン
3. **Pattern 3**: BadgerDBを使用した永続化機能付きブロックチェーン

## セットアップ

### 必要な環境
- Go 1.18以上
- Git

### インストール

1. リポジトリをクローン
```bash
git clone <repository-url>
cd go-blockchain-app
```

2. 依存関係をインストール
```bash
go mod tidy
```

## 使用方法

### 基本コマンド

```bash
go run *.go <pattern>
```

- `<pattern>`: 1, 2, 3 のいずれか

### Pattern 1: シンプルなブロックチェーン

最も基本的なブロックチェーン実装です。メモリ上でブロックを管理し、SHA-256ハッシュを使用してチェーンを構築します。

```bash
# 実行
go run *.go 1
```

**特徴:**
- メモリベースの配列でブロックを管理
- 基本的なハッシュ計算
- プログラム終了時にデータは消失

**出力例:**
```
=== Simple Blockchain (Pattern 1) ===
Timestamp: 1754139553
Data: Genesis Block
PrevBlockHash: 
Hash: 5d10b0d748c4f5285909946dd221950a50cc5ea8116eae56858c45098dd8a34e

Timestamp: 1754139553
Data: Send 1 BTC to Alice
PrevBlockHash: 5d10b0d748c4f5285909946dd221950a50cc5ea8116eae56858c45098dd8a34e
Hash: 80192f4f90a025e47fb5423717bf63eff92397498b0012abe6f0e35381b618ab
```

### Pattern 2: Proof of Work（PoW）機能付きブロックチェーン

マイニング機能を持つブロックチェーンです。計算量による合意メカニズムを実装しています。

```bash
# 実行
go run *.go 2
```

**特徴:**
- Proof of Workアルゴリズム（難易度24）
- ナンス（Nonce）による計算競争
- マイニング時間の可視化
- PoW検証機能

**出力例:**
```
=== Blockchain with Proof of Work (Pattern 2) ===
Mining the block containing "Genesis Block"
0000003f3ae6a7c751c9f5ca04065683b9e1ca244d01bc387b384751c48b8738

Timestamp: 1754139861
Data: Genesis Block
Hash: 0000003f3ae6a7c751c9f5ca04065683b9e1ca244d01bc387b384751c48b8738
Nonce: 1375837
Difficulty: 24
PoW: true
```

### Pattern 3: BadgerDBを使用した永続化機能付きブロックチェーン

データベースでブロックを永続化し、CLI機能を持つ本格的なブロックチェーンです。

#### ブロックの追加

```bash
# データ付きブロックを追加
go run *.go 3 add "Send 1 BTC to Alice"
go run *.go 3 add "Send 2 BTC to Bob"
```

#### 全ブロックの表示

```bash
# 全ブロックを表示（最新から古い順）
go run *.go 3 print
```

**特徴:**
- BadgerDBによるデータ永続化
- encoding/gobによるシリアライズ
- BlockchainIteratorによる効率的なデータ取得
- CLI機能（add/print）
- プログラム再起動後もデータが保持

**使用例:**

1. ブロック追加:
```bash
$ go run *.go 3 add "Send 1 BTC to Alice"
=== Blockchain with Persistent Storage (Pattern 3) ===
Mining the block containing "Send 1 BTC to Alice"
00009022d499010e2954255c822dfccaa82a0dfff1013d813c05cc8469d95b61
Block added: Send 1 BTC to Alice
```

2. データ表示:
```bash
$ go run *.go 3 print
=== Blockchain with Persistent Storage (Pattern 3) ===
============ Block 00009022... ============
Timestamp: 1754140300
Data: Send 1 BTC to Alice
Hash: 00009022d499010e2954255c822dfccaa82a0dfff1013d813c05cc8469d95b61
Nonce: 130662
Difficulty: 16
PoW: true
```

## ファイル構成

```
├── main.go              # メインエントリーポイント
├── blockchain-one.go    # Pattern 1: シンプルブロックチェーン
├── blockchain-two.go    # Pattern 2: PoW付きブロックチェーン
├── blockchain-three.go  # Pattern 3: 永続化ブロックチェーン
├── go.mod              # Go modules設定
├── go.sum              # 依存関係のハッシュ
├── blockchain.db/      # BadgerDBデータファイル（Pattern 3使用時に作成）
└── README.md           # このファイル
```

## 技術的詳細

### Pattern 1の実装
- **構造体**: `Block`, `Blockchain`
- **ハッシュ**: SHA-256
- **データ管理**: メモリ上の配列

### Pattern 2の実装
- **構造体**: `BlockPoW`, `BlockchainPoW`, `ProofOfWork`
- **マイニング**: target値との比較によるナンス探索
- **難易度**: 24（先頭24ビットが0になるハッシュを探索）

### Pattern 3の実装
- **データベース**: BadgerDB
- **シリアライズ**: encoding/gob
- **キー管理**: `lh`（last hash）による最新ブロック追跡
- **イテレータ**: 効率的なブロックチェーン走査

## 学習ポイント

1. **ブロックチェーンの基本概念**
   - ブロック構造
   - ハッシュチェーン
   - 前ブロックへの参照

2. **Proof of Work**
   - マイニングプロセス
   - 難易度調整
   - 計算量による合意

3. **データ永続化**
   - NoSQLデータベース（BadgerDB）
   - シリアライゼーション
   - トランザクション管理

## トラブルシューティング

### よくある問題

1. **BadgerDBのロックエラー**
   - 既存のプロセスを終了してから再実行
   - `blockchain.db`ディレクトリの削除

2. **モジュールエラー**
   - `go mod tidy`を実行
   - Go 1.18以上を使用

3. **メモリ不足**
   - Pattern 2で長時間マイニングが続く場合、難易度を下げる
   - `targetBits`の値を小さくする（例: 16）

## 今後の拡張案

- トランザクション機能
- デジタル署名
- P2Pネットワーク
- スマートコントラクト
- 異なるコンセンサスアルゴリズム（Proof of Stake等）

## ライセンス

MIT License