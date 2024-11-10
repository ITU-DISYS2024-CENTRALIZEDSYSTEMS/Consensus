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
	Timestamp  int32
	Identifier int
	Peers      []string
	Requesting sync.Mutex
}

// Finds the first available port from an array of ports in .env file. Then returns an listener on that port.
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
	log.Printf("Accessing critical resource.")
	time.Sleep(5 * time.Second)
	log.Println("Releasing critical resource.")
}

func getAccessToResource(server *ConsensusServer) {
	grpcOptions := grpc.WithTransportCredentials(insecure.NewCredentials())

	time.Sleep(5 * time.Second) // Timeout before trying to access CS.

	for {
		server.Requesting.Lock()
		accessGranted := true

		log.Println("Requesting access")

		for _, peer := range server.Peers {
			conn, err := grpc.NewClient(":"+peer, grpcOptions)
			if err != nil {
				log.Fatalf("Failed to connect to peer: %s %s", peer, err)
			}

			client := consensus.NewConsensusClient(conn)
			ctx := context.Background()
			reply, err := client.RequestAccess(ctx, &consensus.Request{
				Timestamp:  server.Timestamp,
				Identifier: int32(server.Identifier),
			})
			conn.Close()
			if err != nil {
				log.Printf("Failed to request access from peer %s: %s", peer, err)
				accessGranted = false
				break
			}
			if !reply.Access {
				accessGranted = false
				break
			}
		}

		// assess if access is granted
		if accessGranted {
			useCriticalResource()
			server.Requesting.Unlock()
			time.Sleep(1 * time.Second)
			server.Timestamp++
		} else {
			log.Println("Access denied, retrying...")
			server.Requesting.Unlock()
			time.Sleep(time.Duration(rand.IntN(1000)) * time.Millisecond)
		}
	}
}

func (server *ConsensusServer) RequestAccess(context context.Context, request *consensus.Request) (*consensus.Reply, error) {
	server.Requesting.Lock()
	defer server.Requesting.Unlock()

	access := false
	if request.Timestamp < server.Timestamp || (request.Timestamp == server.Timestamp && request.Identifier > int32(server.Identifier)) {
		access = true
	}

	log.Printf("Peer requested access | ID: %d, TS: %d -> Granted: %v", request.Identifier, request.Timestamp, access)
	return &consensus.Reply{Access: access}, nil
}

func main() {
	// Read ports from .env file
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
		Timestamp:  0,
		Identifier: os.Getpid(),
		Peers:      append(ports[:usedPortIndex], ports[usedPortIndex+1:]...),
	}
	consensus.RegisterConsensusServer(server, service)

	go getAccessToResource(service)

	err = server.Serve(listener)
	if err != nil {
		log.Fatalln("Cannot serve service:", err)
	}
}
