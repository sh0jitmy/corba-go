# corba-go

*他言語で読む: [English](README.md)*

Go言語によるネイティブなCORBAミドルウェアの実装です。このプロジェクトは、完全にGoで構築された基本的なオブジェクト・リクエスト・ブローカー（ORB）とIDLコンパイラ（`idlc`）を提供し、従来のCORBAシステムとの相互運用を目指しています。

## 主な機能
- **IDLコンパイラ (`idlc`)**: OMG IDLをGoのインターフェース、クライアントスタブ、サーバスケルトンにコンパイルします。
  - 対応済みのIDL型: `struct`, `exception`, `typedef`, `sequence`, `enum`, `union`, 基本型（`long`, `short`, `octet`, `string`, `boolean` など）, `Object`, `any`。
  - 生成コードは3つのファイルに分離: `*_types.go`（型定義）、`*_stub.go`（クライアントスタブ）、`*_skel.go`（サーバスケルトン）。
- **ORBコア**: 
  - メモリアライメントを厳密に処理するCDR（Common Data Representation）シリアライズ処理。
  - GIOP 1.2 および IIOP（Internet Inter-ORB Protocol）のサポート。
  - IOR（Interoperable Object Reference）のエンコード/デコードユーティリティ。
- **ネーミングサービス (`CosNaming`)**: インメモリでの名前とIORのバインディング・名前解決。
- **イベントサービス (`CosEvent`)**: サプライヤとコンシューマ間でメッセージをブロードキャストするPushモデルイベントチャネル。

## ディレクトリ構成

```
corba-go/
├── cmd/                          # 各種実行バイナリのメインコード
│   ├── idlc/                     # IDLコンパイラ
│   ├── naming-service/           # ネーミングサービス
│   └── event-service/            # イベントサービス
├── orb/                          # ORBプロトコルのコア機能
│   ├── cdr/                      # CDRエンコーダ/デコーダ + Any型
│   ├── giop/                     # GIOPメッセージ処理
│   ├── iiop/                     # IIOPクライアント/サーバ
│   └── iop/                      # IORエンコード/デコード
├── idlc/                         # IDLコンパイラ内部（字句解析器、構文解析器、AST、コード生成）
├── services/                     # 標準CORBAサービスの実装パッケージ
│   ├── naming/                   # CosNaming（IDL + サーバ実装 + 生成コード）
│   └── event/                    # CosEvent（IDL + サーバ実装 + 生成コード）
└── examples/                     # 使用例
    ├── basic_test/               # 基本的なIIOPクライアント/サーバ
    ├── naming_test/              # ネーミングサービス統合テスト
    │   └── client/               # ネーミングクライアント（名前解決 + 呼び出し）
    └── event_test/               # イベントサービス統合テスト
        ├── supplier/             # Pushサプライヤ
        └── consumer/             # Pushコンシューマ
```

## クイックスタート (Makefile)

```bash
# IDLコンパイラ(idlc)および全サービスのビルド
make build

# サービスIDLファイルからGoコードを生成
make generate

# 生成されたバイナリやコードのクリーンアップ
make clean

# ネーミングサービスの起動
make run-naming

# イベントサービスの起動
make run-event
```

## 手動での実行方法

### 1. IDLのコンパイル

`idlc` コンパイラは各IDL定義から3つのファイルを生成します：

```bash
# コンパイラのビルド
go build -o bin/idlc ./cmd/idlc/

# IDLファイルからGoコードを生成
# 出力: MyService_types.go, MyService_stub.go, MyService_skel.go
./bin/idlc -pkg mypkg MyService.idl
```

**オプション:**
| フラグ | デフォルト | 説明 |
|--------|-----------|------|
| `-pkg` | IDLモジュール名の小文字 | 生成コードのGoパッケージ名 |
| `-out` | `.` | 生成ファイルの出力先ディレクトリ |

### 2. サービスの起動

```bash
# ネーミングサービスの起動 (デフォルトポート :2809)
go run cmd/naming-service/main.go

# イベントサービスの起動 (デフォルトポート :2810)
go run cmd/event-service/main.go
```

### 3. IORユーティリティ

`iop` パッケージを使用してIOR文字列の作成・解析ができます：

```go
import "github.com/sh0jitmy/corba-go/orb/iop"

// IORの作成
ior := iop.NewIOR("IDL:MyModule/MyInterface:1.0", "localhost", 2809, []byte("MyObject"))
iorStr := ior.StringifyIOR()

// IOR文字列の解析
parsed, _ := iop.ParseIOR(iorStr)
profile, _ := iop.ParseIIOPProfile(parsed.Profiles[0].ProfileData)
fmt.Printf("Host: %s, Port: %d, ObjectKey: %s\n", profile.Host, profile.Port, string(profile.ObjectKey))
```

## サンプルコード

### ネーミングサービステスト

CosNamingを使った名前登録と解決のデモ：

```bash
# ターミナル1: ネーミングサービスを起動
./bin/naming-service

# ターミナル2: サーバを起動してネーミングサービスに登録
cd examples/naming_test
go run server_main.go test_types.go test_stub.go test_skel.go

# ターミナル3: 名前を解決してオペレーションを呼び出し
cd examples/naming_test/client
go run client_main.go test_types.go test_stub.go test_skel.go
```

### イベントサービステスト

CosEventを使ったPushモデルのイベントブロードキャストのデモ：

```bash
# ターミナル1: イベントサービスを起動
./bin/event-service

# ターミナル2: コンシューマを登録
cd examples/event_test/consumer
go run consumer_main.go

# ターミナル3: サプライヤからイベントをPush
cd examples/event_test/supplier
go run supplier_main.go
```

サプライヤは `cdr.Any` 型のイベントをEventChannelを通じてPushし、サービスが登録済みの全コンシューマにブロードキャストします。

## ライセンス
このプロジェクトは [Apache 2.0 License](LICENSE) で公開されています。
