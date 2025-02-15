# Stage 1: ビルド環境
FROM golang:1.23-alpine AS builder
WORKDIR /app

# 依存関係のダウンロード
COPY go.mod go.sum ./
RUN go mod download

# ソースコードをコピー
COPY . .

# サーバーとクライアントのビルド
# それぞれの main.go をビルドして、実行ファイルを生成します。
RUN CGO_ENABLED=0 go build -o server ./cmd/server/main.go
RUN CGO_ENABLED=0 go build -o client ./cmd/client/main.go

# Stage 2: 実行環境
FROM alpine:latest
WORKDIR /root/

# ビルド済みバイナリをコピー
COPY --from=builder /app/server .
COPY --from=builder /app/client .

# サーバーで使用するポートを公開
EXPOSE 50051

# デフォルトのコマンドはサーバーを起動する
CMD ["./server"]
