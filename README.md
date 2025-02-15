# Go gRPC サンプルアプリケーション(Greetly)

## 概要

このアプリケーションは、Go と gRPC を使ったシンプルなコードです。  
サーバーは、クライアントから送られる名前情報を受け取り、 `"Hello [Name]"` のレスポンスを返します。  
また、Docker および docker-compose を利用して、サーバーとクライアントをコンテナ化し、環境に応じた接続先設定（ホスト環境とコンテナ内環境の切り替え）を行っています。

## 特徴

- **gRPC サーバー/クライアントの実装**  
  Protocol Buffers (.proto) ファイルから自動生成されたコード（hello.pb.go, hello_grpc.pb.go）を利用しています。

- **Docker 化**  
  サーバーとクライアントを個別のコンテナとして起動し、docker-compose による統合管理を実現。

- **環境ごとの接続先切替**  
  ホスト環境では `localhost:50052`、コンテナ内では `server:50051` を使って接続先を切り替えます。

## ディレクトリ構成

```
.
├── cmd
│   ├── client
│   │   └── main.go         # gRPC クライアントのエントリーポイント
│   └── server
│       └── main.go         # gRPC サーバーのエントリーポイント
├── proto
│   └── hello
│       ├── hello.proto     # サービスとメッセージの定義
│       ├── hello.pb.go     # 自動生成されたメッセージ型のコード
│       └── hello_grpc.pb.go# 自動生成された gRPC サービスのコード
├── go.mod
├── go.sum
└── README.md
```

## セットアップ

### ローカル環境での準備

1. **Go Modules の依存関係を取得**

   プロジェクトルートで以下を実行します。

   ```bash
   go mod tidy
   ```

2. **Protocol Buffers のコード生成**

   `hello.proto` は `proto/hello/` に配置している前提です。  
   プロジェクトルートから以下のコマンドを実行してください。

   ```bash
   protoc --go_out=paths=source_relative:. --go-grpc_out=paths=source_relative:. proto/hello/hello.proto
   ```

   これにより、`proto/hello/` 内に自動生成コード（hello.pb.go と hello_grpc.pb.go）が作成されます。

### Docker 環境での開発

このプロジェクトでは、Dockerfile と docker-compose.yml を使って、サーバーとクライアントをコンテナ化しています。

#### Dockerfile の例

```dockerfile
# Stage 1: ビルド環境
FROM golang:1.23-alpine AS builder
WORKDIR /app

# Go Modules の依存関係を取得
COPY go.mod go.sum ./
RUN go mod download

# ソースコードをコピー
COPY . .

# サーバーとクライアントのビルド（CGO_ENABLED=0 により静的バイナリを生成）
RUN CGO_ENABLED=0 go build -o server ./cmd/server/main.go
RUN CGO_ENABLED=0 go build -o client ./cmd/client/main.go

# Stage 2: 実行環境
FROM alpine:latest
WORKDIR /root/

# ビルド済みバイナリをコピー
COPY --from=builder /app/server .
COPY --from=builder /app/client .

# サーバーで使用するポート（コンテナ内部は 50051）
EXPOSE 50051

# デフォルトのコマンドはサーバーを起動
CMD ["./server"]
```

#### docker-compose.yml の例

```yaml
version: '3.8'

services:
  server:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - '50052:50051' # ホストの 50052 をコンテナの 50051 にマッピング
    command: ['./server']

  client:
    build:
      context: .
      dockerfile: Dockerfile
    depends_on:
      - server
    environment:
      - GRPC_TARGET=server:50051 # クライアントはコンテナ内部から server に接続
    command: ['./client']
```

#### 接続先の切替

クライアントコード内で、環境変数 `GRPC_TARGET` を使って接続先を切り替える実装例:

```go
target := os.Getenv("GRPC_TARGET")
if target == "" {
    // 環境変数が設定されていなければ、ホスト向けのデフォルト値を使用
    target = "localhost:50052"
}
conn, err := grpc.NewClient("passthrough:///" + target,
    grpc.WithTransportCredentials(insecure.NewCredentials()),
)
// 以下、接続開始などの処理…
```

## サーバーの実装概要

`proto/hello/hello.proto` で定義したサービスに基づいて、自動生成されたコードには `GreeterServer` インターフェースと、`UnimplementedGreeterServer` が含まれます。  
サーバー側の実装では、以下のように `UnimplementedGreeterServer` を埋め込み、必要なメソッド（ここでは `SayHello`）だけを実装します。

### server/main.go の例

```go
package main

import (
    "context"
    "log"
    "net"

    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"
    pb "github.com/yourusername/yourproject/proto/hello"
)

// server 構造体は Greeter サービスの実装を行う型です。
// UnimplementedGreeterServer を埋め込むことで、未実装メソッドのデフォルト実装を利用します。
type server struct {
    pb.UnimplementedGreeterServer
}

// SayHello は、Greeter サービスで定義された RPC メソッドです。
// クライアントから送られる HelloRequest を受け取り、HelloResponse を返します。
func (s *server) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
    log.Printf("Received request for name: %s", req.GetName())
    return &pb.HelloResponse{Message: "Hello " + req.GetName()}, nil
}

func main() {
    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }

    s := grpc.NewServer()
    pb.RegisterGreeterServer(s, &server{})

    // Reflection を有効にして、外部ツール（例：grpcurl）がサービス情報を取得できるようにする
    reflection.Register(s)

    log.Printf("Server listening at %v", lis.Addr())
    if err := s.Serve(lis); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}
```

### ポイント

- **hello.proto の定義に基づく自動生成コード:**
  - `GreeterServer` インターフェースと `UnimplementedGreeterServer` が含まれており、これを実装することで、サーバー側の RPC メソッドを定義します。
- **SayHello の実装:**
  - クライアントからのリクエストで受け取った `name` を取得し、ログ出力。
  - レスポンスとして `"Hello " + req.GetName()` を返します。

## サーバーとクライアントの起動

### ローカルでの起動

- サーバー:
  ```bash
  go run cmd/server/main.go
  ```
- クライアント:
  ```bash
  go run cmd/client/main.go
  ```

### Docker での起動

プロジェクトルートで以下のコマンドを実行してください。

```bash
docker-compose up --build
```

- サーバーコンテナはホストのポート 50052 にマッピングされます。
- クライアントは環境変数 `GRPC_TARGET` に基づいてサーバーへ接続します。
