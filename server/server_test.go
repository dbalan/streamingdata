package main

import (
	"fmt"
	pb "github.com/dbalan/streamingdata/streamingdata"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
	"io"
	// 	"io/ioutil"
	"os/exec"
	"testing"
	"time"
)

var testport int = 8000
var conn *grpc.ClientConn
var client pb.RealTimeClient

func setup() {
	var err error
	go func() {
		cmd := exec.Command("./server", "-port", fmt.Sprintf("%d", testport))
		err := cmd.Start()
		if err != nil {
			panic("starting server failed")
		}
	}()

	time.Sleep(1 * time.Second)

	conn, err = grpc.Dial(fmt.Sprintf("localhost:%d", testport), grpc.WithInsecure())
	if err != nil {
		panic("grpc connection failed")
	}
	client = pb.NewRealTimeClient(conn)

}

func TestGetPing(t *testing.T) {
	// well, this is hacky
	setup()
	resp, err := client.GetPing(context.Background(), &pb.Empty{})
	if err != nil {
		t.Error(err)
	}
	if resp.Resp != "PONG" {
		t.Error("invalid response")
	}
}

func TestStateLessApi(t *testing.T) {
	req := &pb.StateLessRequest{Count: 3}
	stream, err := client.GetStateLessStream(context.Background(), req)
	if err != nil {
		t.Error("stateless query returned err", err)
	}

	for {
		_, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Error("error while reading stream", err)
		}
	}
}

func TestStateLessApiReconnect(t *testing.T) {
	req := &pb.StateLessRequest{Count: 3}
	stream, err := client.GetStateLessStream(context.Background(), req)
	if err != nil {
		t.Error("stateless query returned err", err)
	}
	resp, err := stream.Recv()
	if err != nil {
		t.Error("something wrong happened")
	}
	req.Count = req.Count - 1
	req.Lastseen = resp.CurrentVal

	// try again.
	stream, err = client.GetStateLessStream(context.Background(), req)
	if err != nil {
		t.Error("stateless query reconnect returned err", err)
	}

	for {
		_, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Error("error while reading stream", err)
		}

	}
}

func TestStateFullApi(t *testing.T) {
	var hash string
	req := &pb.StateFullRequest{Clientid: "dummyclient22", Count: 3}
	stream, err := client.GetStateFullStream(context.Background(), req)
	count := 0
	if err != nil {
		t.Error("statefull stream failed!")
	}

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Error("statefull failed with err: ", err)
		}
		hash = resp.HashSum
		count++
	}

	if count != 3 {
		t.Error("got unspecified number of responses")
	}
	if hash == "" {
		t.Error("hash value is not send with last message")
	}

}

// Failing because I need to figure out how to simulate client connection drops
/*
func TestStateFullApiReconnect(t *testing.T) {
	var hash string
	req := &pb.StateFullRequest{Clientid: "dummyclient", Count: 3}
	stream, err := client.GetStateFullStream(context.Background(), req)
	count := 0
	if err != nil {
		t.Error("statefull stream failed!")
	}

	_, err = stream.Recv()
	if err != nil {
		t.Error("reciving failed", err)
	}

	_ = conn.Close()
	conn, err = grpc.Dial(fmt.Sprintf("localhost:%d", testport), grpc.WithInsecure())
	client = pb.NewRealTimeClient(conn)
	stream, err = client.GetStateFullStream(context.Background(), req)

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Error("statefull failed with err: ", err)
		}
		hash = resp.HashSum
		count++
	}

	if count != 2 {
		t.Error("got unspecified number of responses", count)
	}
	if hash == "" {
		t.Error("hash value is not send with last message")
	}

}
*/
