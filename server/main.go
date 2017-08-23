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
	for i = 0; i < req.Count; i++ {
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
	var i int64

	fmt.Println("trying to retrive things")
	session, err = storage.Retrive(req.Clientid)
	if err != nil {
		if err == KeyExpiredError {
			log.Printf("key expired: %v", err)
			return err
		} else if err == AlreadyPresentError {
			log.Printf("session in progress, please try again")
			return err
		}
		// assume new session
		session = &SessionData{
			ClientID: req.Clientid,
			Seed:     rand.Int63(),
		}

	}

	h := sha256.New()
	rng := rand.New(rand.NewSource(session.Seed))
	if session.Last != 0 {
		for i = 0; i < session.Last; i++ {
			cur := rng.Uint32()
			fmt.Printf("replying: %d\n", cur)
			h.Write([]byte(fmt.Sprintf("%d", cur)))
		}

	}

	mtosnd := req.Count - session.Last
	for i = 0; int64(i) < mtosnd; i++ {
		cur := rng.Uint32()
		h.Write([]byte(fmt.Sprintf("%d", cur)))

		resp := &pb.StateFullResponse{CurrentVal: cur}

		// FIXME: there is a cleaner way to do this.
		if i == mtosnd-1 {
			resp.HashSum = fmt.Sprintf("%x", h.Sum(nil))
		}

		// only for testing by killing server.
		/*
			session.DiscardTs = time.Now().Unix()
			session.Last = i + 1
			err = storage.Store(session)
			if err != nil {
				log.Fatal("storage failed!")
			}
		*/
		if err = stream.Send(resp); err != nil {
			// Sending failed.
			session.Last = i
			session.DiscardTs = time.Now().Unix()
			err = storage.Store(session)
			if err != nil {
				log.Fatal("storage backend failed")
			}
			break
		}
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
