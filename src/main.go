package main

import (
	"context"
	"errors"
	"log"
	"math/rand/v2"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"consensus/consensus"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ConsensusServer struct {
	consensus.UnimplementedConsensusServer
	timestamp  int32
	identifier int
	peers      []string
	requesting sync.Mutex
	accessing  sync.Mutex
}

// Finds the first available port from an array of ports. Then returns an listener on that port.
func findAvailablePort(ports []string) (listener net.Listener, err error) {
	for _, port := range ports {
		listener, err := net.Listen("tcp", ":"+port)
		if err == nil {
			return listener, nil
		}
	}

	return nil, err
}

// Returns the index where the current port used by the listener is, in the given array.
func selectedPortIndex(ports []string, listener net.Listener) (index int, err error) {
	selectedPort := strconv.Itoa(listener.Addr().(*net.TCPAddr).Port)

	for i, port := range ports {
		if selectedPort == port {
			return i, nil
		}
	}

	return -1, errors.New("port not found")
}

// Emulates a resource being accessed.
func useCriticalResource() {
	log.Printf("I got access the resource!")
	log.Println("Waiting 5 secounds to ask again...")
	time.Sleep(5 * time.Second)
}

func getAccessToResource(server *ConsensusServer) {
	grpcOptions := grpc.WithTransportCredentials(insecure.NewCredentials())

	for {
		accessGranted := true
		for _, peer := range server.peers {
			conn, err := grpc.NewClient(":"+peer, grpcOptions)
			if err != nil {
				log.Fatalf("Failed to connect to peer: %s %s", peer, err)
			}
			defer conn.Close()

			client := consensus.NewConsensusClient(conn)
			ctx := context.Background()
			reply, err := client.RequestAccess(ctx, &consensus.Request{
				Timestamp:  server.timestamp,
				Identifier: int32(server.identifier),
			})
			if err != nil || !reply.Access {
				accessGranted = false
				break
			}
		}

		if accessGranted {
			useCriticalResource()
		} else {
			log.Println("Access denied, retrying...")
			server.timestamp++
			time.Sleep(time.Duration(rand.IntN(1000)) * time.Millisecond)
		}
	}
}

func (server *ConsensusServer) RequestAccess(context context.Context, request *consensus.Request) (*consensus.Reply, error) {
	server.requesting.Lock()
	defer server.requesting.Unlock()

	access := request.Timestamp >= server.timestamp

	return &consensus.Reply{Access: access}, nil
}

func main() {
	envFile, _ := godotenv.Read(".env")
	ports := strings.Split(envFile["PORTS"], ",")

	listener, err := findAvailablePort(ports)
	if err != nil {
		log.Fatalln("Cannot create listener:", err)
	}

	log.Println("Using port:", listener.Addr().String())

	usedPortIndex, err := selectedPortIndex(ports, listener)
	if err != nil {
		log.Fatalln("Cannot find used port", err)
	}

	server := grpc.NewServer()
	service := &ConsensusServer{
		timestamp:  0,
		identifier: os.Getpid(),
		peers:      append(ports[:usedPortIndex], ports[usedPortIndex+1:]...),
	}
	consensus.RegisterConsensusServer(server, service)

	go getAccessToResource(service)

	err = server.Serve(listener)
	if err != nil {
		log.Fatalln("Cannot serve service:", err)
	}
}
