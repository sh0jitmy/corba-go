# corba-go

*他言語で読む: [English](README.md)*

Go言語によるネイティブなCORBAミドルウェアの実装です。このプロジェクトは、完全にGoで構築された基本的なオブジェクト・リクエスト・ブローカー（ORB）とIDLコンパイラ（`idlc`）を提供し、レガシーなCORBAシステム（例：omniORB、JacORB）との相互運用を可能にします。

## 主な機能
- **IDLコンパイラ (`idlc`)**: OMG IDLをGoのインターフェース、クライアントスタブ、サーバスケルトンにコンパイルします。
  - 対応済みのIDL型: `struct`, `exception`, `typedef`, `sequence`, `enum`, `union`, 基本型（`long`, `short`, `octet`, `string`, `boolean` など）, `Object`。
- **ORBコア**: 
  - メモリアライメントを厳密に処理するCDR（Common Data Representation）シリアライズ処理。
  - GIOP 1.2 および IIOP（Internet Inter-ORB Protocol）のサポート。
- **ネーミングサービス (`CosNaming`)**: インメモリでの名前とIORのバインディング・名前解決。
- **イベントサービス (`CosEvent`)**: メッセージをブロードキャストするためのシンプルなPushモデルイベントチャネル。

## ディレクトリ構成
- `cmd/`: 各種実行バイナリのメインコード。
  - `idlc/`: IDLコンパイラ本体。
  - `naming-service/`: 独立したネーミングサービス。
  - `event-service/`: 独立したイベントサービス。
- `orb/`: ORBプロトコルのコア機能（`cdr`, `giop`, `iiop`, `iop`）。
- `services/`: 標準CORBAサービスの実装パッケージ（`naming`, `event`）。
- `idlc/`: IDLコンパイラのためのパーサー、抽象構文木（AST）、Goコード生成ロジック。

## クイックスタート (Makefile)

ビルドや実行を簡単に行うための `Makefile` を提供しています。

```bash
# IDLコンパイラ(idlc)および全サービスのビルド
make build

# 生成されたバイナリやスタブのクリーンアップ
make clean

# ネーミングサービスの起動
make run-naming

# イベントサービスの起動
make run-event
```

## 手動での実行方法

### 1. IDLのコンパイル
```bash
# コンパイラのビルド
go build -o bin/idlc cmd/idlc/main.go

# IDLファイルからGoコードを生成
./bin/idlc -pkg mypkg my_interface.idl
```

### 2. サービスの起動
```bash
# ネーミングサービスの起動 (デフォルトポート :2809)
go run cmd/naming-service/main.go

# イベントサービスの起動 (デフォルトポート :2810)
go run cmd/event-service/main.go
```
