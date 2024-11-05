package main

import (
	consensus "consensus/consensus"
	"context"
	"log"
	"math/rand/v2"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Peer struct {
	name string
}

type ConsensusServer struct {
	consensus.UnimplementedConsensusServer
	timestamp  int32
	peers      []string
	requesting bool
	mu         sync.Mutex
}

func findAvailablePort(ports []string) (listener net.Listener, err error) {
	for i := 0; i < len(ports); i++ {
		listener, err := net.Listen("tcp", ":"+ports[i])
		if err != nil {
			continue
		}

		return listener, err
	}

	return
}

func selectedPortIndex(ports []string, listener net.Listener) int {
	selectedPort := strconv.Itoa(listener.Addr().(*net.TCPAddr).Port)

	for i := 0; i < len(ports); i++ {
		if selectedPort == ports[i] {
			return i
		}
	}

	return -1
}

func useCriticalResource() {
	log.Printf("I got access the resource!")
	log.Println("Waiting 5 secounds to ask again...")
	time.Sleep(5 * time.Second)
}

func getAccessToResource(server *ConsensusServer) {
	server.mu.Lock()

	time.Sleep(5 * time.Second)

	grpcOptions := grpc.WithTransportCredentials(insecure.NewCredentials())

	for {
		accessGranted := false
		for i := 0; i < len(server.peers); i++ {
			conn, err := grpc.NewClient(":"+server.peers[i], grpcOptions)
			if err != nil {
				log.Fatalf("Cannot create client: %s", err)
			}

			client := consensus.NewConsensusClient(conn)
			ctx := context.Background()
			reply, err := client.RequestAccess(ctx, &consensus.Request{
				Timestamp: server.timestamp,
			})
			if err != nil {
				log.Fatalf("Cannot request access: %s", err)
			}

			accessGranted = reply.Access
		}

		if accessGranted {
			useCriticalResource()
		} else {
			log.Println("Did not get access, waiting...")
			server.timestamp++
			time.Sleep(time.Duration(rand.IntN(1000)) * time.Millisecond)
		}
	}
}

func (server *ConsensusServer) RequestAccess(context context.Context, request *consensus.Request) (*consensus.Reply, error) {
	server.mu.Lock()
	if request.Timestamp <= server.timestamp {
		replyAccess = false
	}

	return &consensus.Reply{}, nil
}

func main() {
	envFile, _ := godotenv.Read(".env")
	ports := strings.Split(envFile["PORTS"], ",")

	listener, err := findAvailablePort(ports)
	if err != nil {
		log.Fatalln("Cannot create listener: ", err)
	}

	log.Println("Using port:", listener.Addr().String())

	usedPortIndex := selectedPortIndex(ports, listener)
	server := grpc.NewServer()
	service := &ConsensusServer{
		timestamp:  0,
		peers:      append(ports[:usedPortIndex], ports[usedPortIndex+1:]...),
		requesting: false,
	}
	consensus.RegisterConsensusServer(server, service)

	go getAccessToResource(service)

	err = server.Serve(listener)
	if err != nil {
		log.Fatalln("Cannot serve service:", err)
	}
}
