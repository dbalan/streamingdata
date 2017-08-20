//go:generate protoc -I ../streamingdata --go_out=plugins=grpc:../streamingdata ../streamingdata/streamingdata.proto
// FIXME: write doc
package main

import (
	"crypto/sha256"
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

	// as long as this interface is satisfied, plug any storage
	storage Storage
)

type realTimeServer struct{}

func (rt *realTimeServer) GetPing(ctx context.Context, e *pb.Empty) (*pb.PingResponse, error) {
	return &pb.PingResponse{"PONG"}, nil
}

func (rt *realTimeServer) GetStateLessStream(req *pb.StateLessRequest,
	stream pb.RealTime_GetStateLessStreamServer) error {

	if req.Count <= 0 {
		return fmt.Errorf("empty count")
	}

	var start uint32 = req.Lastseen
	for start == 0 {
		// FIXME: where is this seeded from?
		start = rand.Uint32()
	}

	var i int64 = 0
	for {
		if i >= req.Count {
			break
		}
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
		i++
	}
	return nil
}

func (rt *realTimeServer) GetStateFullStream(req *pb.StateFullRequest,
	stream pb.RealTime_GetStateFullStreamServer) error {
	var err error
	var session *SessionData

	if req.Reconnect {
		session, err = storage.Retrive(req.Clientid)
		if err != nil {
			log.Printf("error retriving state: %v", err)
			return err
		}

	} else {
		session = &SessionData{
			ClientID: req.Clientid,
			Count:    req.Count,
			Seed:     rand.Int63(),
		}
	}

	h := sha256.New()
	rng := rand.New(rand.NewSource(session.Seed))
	if req.Count < session.Count {
		// then this is a reconnect
		sentCnt := session.Count - req.Count
		// FIXME: loop can be better
		var i int64 = 0
		for {
			if i >= sentCnt {
				break
			}
			cur := rng.Uint32()
			fmt.Printf("replying: %d", cur)
			h.Write([]byte(fmt.Sprintf("%d", cur)))
			i++
		}

	}

	var i int64 = 0
	for {
		if i >= req.Count {
			break
		}
		cur := rng.Uint32()
		h.Write([]byte(fmt.Sprintf("%d", cur)))

		resp := &pb.StateFullResponse{CurrentVal: cur}

		// FIXME: there is a cleaner way to do this.
		if i == req.Count-1 {
			resp.HashSum = fmt.Sprintf("%x", h.Sum(nil))
		}

		// only for testing by killing server.
		/*
			session.DiscardTs = time.Now().Unix()
			err = storage.Store(session)
			if err != nil {
				log.Fatal("storage failed!")
			}
		*/
		if err = stream.Send(resp); err != nil {
			// Sending failed.
			session.DiscardTs = time.Now().Unix()
			err = storage.Store(session)
			if err != nil {
				log.Fatal("storage backend failed")
			}
			break
		}
		i++
		time.Sleep(1 * time.Second)
	}
	return nil
}

func main() {
	flag.Parse()
	storage = NewInMemStorage()
	// storage = NewRedisStorage()
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
