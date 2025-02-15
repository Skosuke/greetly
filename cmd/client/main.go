package main

import (
	"context"
	"log"
	"time"

	pb "github.com/Skosuke/greetly/proto/hello"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// NewClientで接続を作成（passthroughでローカル接続）
	conn, err := grpc.NewClient("passthrough:///localhost:50051",
			grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
			log.Fatalf("Failed to create client: %v", err)
	}
	defer conn.Close()

	// 手動で接続を開始（NewClientは接続を遅延する）
	conn.Connect()

	// 接続状態の確認
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for conn.GetState().String() != "READY" {
			if !conn.WaitForStateChange(ctx, conn.GetState()) {
					log.Fatalf("Failed to connect within the timeout")
			}
	}
	log.Println("Client connected successfully")

	// クライアントの作成＆リモートメソッド呼び出し
	client := pb.NewGreeterClient(conn)
	res, err := client.SayHello(context.Background(), &pb.HelloRequest{Name: "Gopher"})
	if err != nil {
			log.Fatalf("Could not greet: %v", err)
	}

	log.Printf("Server response: %s", res.Message)
}