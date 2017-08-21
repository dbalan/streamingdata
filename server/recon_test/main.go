package main

import (
	"flag"
	"fmt"
	pb "github.com/dbalan/streamingdata/streamingdata"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
	"io"
	"os"
)

var (
	fail     = flag.Bool("fail", false, "fail after first request")
	clientid = flag.String("cid", "helloworld", "client id")
	count    = flag.Int64("count", 5, "number of requests")
	port     = flag.Int("port", 8000, "the server port")
)

func main() {
	flag.Parse()
	conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", *port), grpc.WithInsecure())
	if err != nil {
		panic("grpc connection failed")
	}
	client := pb.NewRealTimeClient(conn)

	req := &pb.StateFullRequest{Clientid: *clientid, Count: *count}
	stream, err := client.GetStateFullStream(context.Background(), req)

	if err != nil {
		panic("statefull stream failed!")
	}
	if *fail {
		resp, err := stream.Recv()
		if err != nil {
			panic("error from stream")
		}
		fmt.Println(resp.CurrentVal, *clientid)
		os.Exit(0)
	}

	i := 0
	for {
		_, err = stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic("error receiving from stream")
		}
		i++
	}
	fmt.Println(i)
}
