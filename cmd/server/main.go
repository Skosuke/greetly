package main

import (
	"context"
	"log"
	"net"

	pb "github.com/Skosuke/greetly/proto/hello" // 生成されたコードをインポート
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// server は Greeter サービスの実装です。
type server struct {
    pb.UnimplementedGreeterServer // 新しいプロジェクトでは、この埋め込みが推奨されます
}

// SayHello は Greeter サービスの RPC メソッドの実装です。
func (s *server) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
    log.Printf("Received request for name: %s", req.GetName())
    // シンプルに、"Hello <name>" というメッセージを返す
    return &pb.HelloResponse{Message: "Hello " + req.GetName()}, nil
}

func main() {
    // 50051 ポートでリッスン
    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }

    // gRPC サーバーの作成
    s := grpc.NewServer()
    // 生成された RegisterGreeterServer を使ってサービスを登録
    pb.RegisterGreeterServer(s, &server{})

		// Reflection を登録する
    reflection.Register(s)

    log.Printf("gRPC server listening at %v", lis.Addr())
    if err := s.Serve(lis); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}
