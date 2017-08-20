//go:generate protoc -I ../streamingdata --go_out=plugins=grpc:../streamingdata ../streamingdata/streamingdata.proto
// FIXME: write doc
package main

import (
	"flag"
	"fmt"
	pb "github.com/dbalan/streamingdata/streamingdata"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"net"
)

var (
	port = flag.Int("port", 8000, "The Server Port")
)

type realTimeServer struct{}

func (rt *realTimeServer) GetPing(ctx context.Context, e *pb.Empty) (*pb.PingResponse, error) {
	return &pb.PingResponse{"PONG"}, nil
}

func main() {
	flag.Parse()
	sock, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterRealTimeServer(grpcServer, &realTimeServer{})
	err = grpcServer.Serve(sock)
	if err != nil {
		log.Fatalf("failed to start gprc server: %v", err)
	}
}
