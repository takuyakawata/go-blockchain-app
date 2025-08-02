# Go Blockchain App

このプロジェクトは、Go言語で実装された段階的なブロックチェーン学習プロジェクトです。シンプルなブロックチェーンから本格的なP2Pネットワーク機能まで、7つの異なるパターンで段階的にブロックチェーン技術を学習できます。

## 概要

このプロジェクトでは以下の7つのブロックチェーンパターンを順次学習できます：

1. **Pattern 1**: シンプルなブロックチェーン（メモリベース）
2. **Pattern 2**: Proof of Work（PoW）機能付きブロックチェーン
3. **Pattern 3**: BadgerDBを使用した永続化機能付きブロックチェーン
4. **Pattern 4**: Cobraライブラリを使用したCLIインターフェース
5. **Pattern 5**: ECDSA暗号化によるウォレットシステム
6. **Pattern 6**: トランザクションとUTXO（未使用トランザクション出力）モデル
7. **Pattern 7**: P2Pネットワーク層とブロックチェーン同期

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

- `<pattern>`: 1, 2, 3, 4, 5, 6, 7 のいずれか

---

## Pattern 1: シンプルなブロックチェーン

最も基本的なブロックチェーン実装です。メモリ上でブロックを管理し、SHA-256ハッシュを使用してチェーンを構築します。

```bash
go run *.go 1
```

**特徴:**
- メモリベースの配列でブロックを管理
- 基本的なハッシュ計算
- プログラム終了時にデータは消失

---

## Pattern 2: Proof of Work（PoW）機能付きブロックチェーン

マイニング機能を持つブロックチェーンです。計算量による合意メカニズムを実装しています。

```bash
go run *.go 2
```

**特徴:**
- Proof of Workアルゴリズム（難易度24）
- ナンス（Nonce）による計算競争
- マイニング時間の可視化
- PoW検証機能

---

## Pattern 3: BadgerDBを使用した永続化機能付きブロックチェーン

データベースでブロックを永続化し、CLI機能を持つ本格的なブロックチェーンです。

```bash
# データ付きブロックを追加
go run *.go 3 add "Send 1 BTC to Alice"

# 全ブロックを表示
go run *.go 3 print
```

**特徴:**
- BadgerDBによるデータ永続化
- encoding/gobによるシリアライズ
- BlockchainIteratorによる効率的なデータ取得
- CLI機能（add/print）

---

## Pattern 4: Cobraライブラリを使用したCLIインターフェース

Cobraライブラリを使用してプロフェッショナルなCLIインターフェースを提供します。

```bash
go run *.go 4
```

**利用可能コマンド:**
- `blockchain` - ブロックチェーン操作
- `addblock <data>` - ブロック追加
- `printchain` - チェーン表示

**特徴:**
- Cobraライブラリによる本格的CLI
- サブコマンド機能
- ヘルプ機能とフラグサポート

---

## Pattern 5: ECDSA暗号化によるウォレットシステム

楕円曲線デジタル署名アルゴリズム（ECDSA）を使用したウォレット機能を実装します。

```bash
go run *.go 5
```

**特徴:**
- ECDSA楕円曲線暗号による鍵ペア生成
- SHA-256 + RIPEMD-160によるアドレス生成
- Base58エンコーディング（Bitcoin形式）
- ウォレットファイルの永続化

**生成されるアドレス例:**
```
1A2B3C4D5E6F7G8H9I0J1K2L3M4N5O6P7Q8R9S
```

---

## Pattern 6: トランザクションとUTXO（未使用トランザクション出力）モデル

Bitcoinと同様のUTXOモデルを実装し、本格的なトランザクション機能を提供します。

```bash
go run *.go 6
```

**機能:**
- UTXO（Unspent Transaction Output）モデル
- デジタル署名によるトランザクション認証
- Coinbaseトランザクション（新規コイン生成）
- 複数入力・複数出力のトランザクション
- 残高計算とトランザクション検証

**主要コンポーネント:**
- `Transaction` - トランザクション構造体
- `TXInput` - トランザクション入力
- `TXOutput` - トランザクション出力
- `UTXOSet` - UTXO管理システム

---

## Pattern 7: P2Pネットワーク層とブロックチェーン同期

本格的なP2Pネットワーク機能を実装し、複数ノード間でのブロックチェーン同期を実現します。

```bash
go run *.go 7
```

### インタラクティブコマンド

Pattern 7では以下のコマンドが利用可能です：

**ノード起動:**
```bash
startnode <port> [bootstrap_nodes...]
```

例：
```bash
# 最初のノード（ブートストラップノード）
startnode 3000

# 他のノード（別ターミナルで実行）
startnode 3001 localhost:3000
startnode 3002 localhost:3000 localhost:3001
```

**その他のコマンド:**
- `nodeinfo <port>` - ノード情報取得
- `connectpeer <local_port> <peer_addr>` - ピア接続
- `listpeers <port>` - 接続ピア一覧
- `sendtx <port> <from> <to> <amount>` - トランザクション送信
- `mineblock <port>` - ブロックマイニング
- `syncstatus <port>` - 同期状態確認

**主要機能:**
- **TCP通信**: ノード間のメッセージ交換
- **プロトコル実装**: version, getblocks, inv, getdata, block, tx, ping/pong
- **ブロックチェーン同期**: 最長チェーン原則による自動同期
- **Mempool管理**: 未確認トランザクションの管理と検証
- **ノード管理**: ピア発見、ヘルスチェック、ブートストラップ
- **ネットワーク監視**: リアルタイムネットワーク状態表示

### P2Pネットワーク構成例

```
Node A (3000) ←→ Node B (3001) ←→ Node C (3002)
      ↓                ↓                ↓
   [Blockchain]   [Blockchain]   [Blockchain]
   [Mempool]      [Mempool]      [Mempool]
```

---

## ファイル構成

```
├── main.go                 # メインエントリーポイント
├── blockchain-one.go       # Pattern 1: シンプルブロックチェーン
├── blockchain-two.go       # Pattern 2: PoW付きブロックチェーン
├── blockchain-three.go     # Pattern 3: 永続化ブロックチェーン
├── blockchain-four.go      # Pattern 4: CLI情報表示
├── blockchain-five.go      # Pattern 5: ウォレットデモ
├── blockchain-six.go       # Pattern 6: トランザクションデモ
├── blockchain-seven.go     # Pattern 7: P2Pネットワーク CLI
├── cmd/                    # Cobraコマンド定義（Pattern 4用）
│   ├── root.go
│   ├── blockchain.go
│   ├── addblock.go
│   └── printchain.go
├── wallet/                 # ウォレット機能（Pattern 5以降）
│   └── wallet.go
├── transaction/            # トランザクション機能（Pattern 6以降）
│   ├── blockchain.go
│   ├── transaction.go
│   └── utxo_set.go
├── network/                # P2Pネットワーク機能（Pattern 7）
│   ├── server.go
│   ├── message.go
│   ├── handlers.go
│   ├── sync.go
│   ├── mempool.go
│   └── node.go
├── go.mod                  # Go modules設定
├── go.sum                  # 依存関係のハッシュ
├── blockchain-*.db/        # BadgerDBデータファイル
├── wallet_*.dat           # ウォレットファイル
└── README.md              # このファイル
```

## 技術的詳細

### 暗号化技術
- **ハッシュ**: SHA-256
- **デジタル署名**: ECDSA（P-256曲線）
- **アドレス生成**: SHA-256 + RIPEMD-160 + Base58

### データ管理
- **データベース**: BadgerDB（Key-Value Store）
- **シリアライズ**: encoding/gob
- **ネットワーク**: TCP + カスタムプロトコル

### アーキテクチャ
- **UTXO モデル**: 未使用トランザクション出力による残高管理
- **P2P ネットワーク**: 分散ノード間の自動同期
- **Mempool**: 未確認トランザクションの一時管理

## 学習の進め方

1. **Pattern 1-3**: ブロックチェーンの基本概念を学習
2. **Pattern 4**: CLI開発のベストプラクティスを学習
3. **Pattern 5**: 暗号化技術とウォレットシステムを学習
4. **Pattern 6**: トランザクションとUTXOモデルを理解
5. **Pattern 7**: P2Pネットワークと分散システムを実装

各パターンは前のパターンの知識を基に構築されているため、順次学習することを推奨します。

## トラブルシューティング

### よくある問題

1. **BadgerDBのロックエラー**
   - 既存のプロセスを終了してから再実行
   - データベースディレクトリの削除

2. **ポート使用中エラー（Pattern 7）**
   - 他のポートを使用
   - プロセスの終了: `pkill -f "go run"`

3. **ネットワーク接続エラー**
   - ファイアウォール設定の確認
   - localhostでの接続テスト

## 今後の拡張案

- スマートコントラクト機能
- 異なるコンセンサスアルゴリズム（Proof of Stake等）
- REST API インターフェース
- Web UI ダッシュボード
- クロスチェーン機能

## ライセンス

MIT License