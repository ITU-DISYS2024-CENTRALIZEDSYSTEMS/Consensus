package main

import (
	consensus "consensus/consensus"
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Ricart–Agrawala algorithm
func main() {
   // var lamport int32 = 1

    lis, port := findAvaliablePort()
    if port == "" && lis == nil {
        log.Fatalf("No available port found")
    }
	defer lis.Close()


	go processPermission(lis)

	//go requestAccessToCS(lamport, port)
}

// func (s *consensus.ConsensusServer) RequestAccess(ctx context.Context, in *consensus.Request) {

// }

func processPermission(lis net.Listener) {
	for{
		grpcServer := grpc.NewServer(grpc.Creds(insecure.NewCredentials()))
		consensus.RegisterConsensusServer(grpcServer, &consensus.UnimplementedConsensusServer{})

		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
		
	}
}

func requestAccessToCS(lamport int32, OwnPort string) {

	for _, port := range getPorts() {
		if port == OwnPort {
			continue
		}
		conn, err := grpc.NewClient(":"+port, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		c := consensus.NewConsensusClient(conn)
		c.RequestAccess(context.Background(), &consensus.Request{Timestamp: lamport})
	}
	
}

func getPorts() []string {
	envFile, err := godotenv.Read(".env")
	if err != nil {
		log.Fatalf("Error reading .env file: %v", err)
	}
	return strings.Split(envFile["PORTS"], ",")
}

func findAvaliablePort() (net.Listener, string) {
    for _, port := range getPorts() {
        lis, err := net.Listen("tcp", ":"+port)
        if err == nil {
            return lis, port
        }
    }
    return nil, ""
}

func wrtieToCS(msg string) {
    filePath := "consensus.txt"
    if _, err := os.Stat(filePath); err == nil {
        // File exists, append to it
        f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
        if err != nil {
            fmt.Println(err)
            return
        }
        defer f.Close()

        l, err := f.WriteString(msg)
        if err != nil {
            fmt.Println(err)
            return
        }
        fmt.Println(l, "bytes written successfully")
    } else if errors.Is(err, os.ErrNotExist) {
        // File does not exist, create it
        f, err := os.Create(filePath)
        if err != nil {
            fmt.Println(err)
            return
        }
        defer f.Close()

        l, err := f.WriteString("Hello World\n")
        if err != nil {
            fmt.Println(err)
            return
        }
        fmt.Println(l, "bytes written successfully")
    } else {
        fmt.Println(err)
    }
}