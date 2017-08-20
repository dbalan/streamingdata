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
	"math/rand"
	"net"
	"time"
)

var (
	port = flag.Int("port", 8000, "The Server Port")
)

type realTimeServer struct{}

func (rt *realTimeServer) GetPing(ctx context.Context, e *pb.Empty) (*pb.PingResponse, error) {
	return &pb.PingResponse{"PONG"}, nil
}

func (rt *realTimeServer) GetStateLessStream(req *pb.StateLessRequest,
	stream pb.RealTime_GetStateLessStreamServer) error {

	var start uint32 = req.Lastseen
	for start == 0 {
		// FIXME: where is this seeded from?
		start = rand.Uint32()
	}

	for i := 0; i < int(req.Count); i++ {
		// we start with twice of the number that rand gave us (or next number
		// in case its a reconnect)
		start = start * 2

		res := &pb.IntResponse{start}

		if err := stream.Send(res); err != nil {
			log.Printf("failed sending: %v", err)
			return err
		}
		// FIXME: small interval for testing
		time.Sleep(1 * time.Second)
	}
	return nil
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
