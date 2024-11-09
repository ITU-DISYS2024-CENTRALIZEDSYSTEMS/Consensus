package main

import (
	"bufio"
	proto "consensus/consensus"
	"context"
	"errors"
	"log"
	"net"
	"os"
	"sync"

	"google.golang.org/grpc"
)


var state = RELEASED
var lamportTime int32 = 0
var usedPort string = ""

var serverList = [3]string{":8080", ":1234", ":4321"}


type consensusServer struct{
	proto.UnimplementedConsensusServer
	lamportTime int32
	mu sync.Mutex
}

func (s *consensusServer) RequestAccess(ctx context.Context, request *proto.Request) (reply *proto.Reply, err error){
	
	rep := &proto.Reply{}

	if (state == HELD || (state == WANTED && (lamportTime < request.Timestamp))){

	}else{
		rep.Access = true
		rep.Timestamp = lamportTime
		log.Println(request.Message)
		return rep, err;
	}
	rep.Access = false
	return rep, nil // Not final
}

func main(){

	port, err := checkAvablePort(serverList[:])
	usedPort = port
	if err != nil {
		log.Fatalln(err)
		return
	}

	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalln("Cannot create listener: ", err)
	}

	server := grpc.NewServer()

	service := &consensusServer{
		lamportTime: 0,
	}

	proto.RegisterConsensusServer(server, service)


	err = server.Serve(listener)
	if err != nil {
		log.Fatalln("Cannot serve service:", err)
	}

}

func SendMessage(){
	for {
		input := bufio.NewScanner(os.Stdin)
		input.Scan()

		if input.Text() == ""{
			continue
		}

		request := &proto.Request{
			Timestamp: lamportTime,
			Message: input.Text(),
		}

		for serverPort := range serverList{
			if serverList[serverPort] == usedPort{
				continue
			}

			go SendRequest(request, serverList[serverPort])
		}
	}
}

func SendRequest(request *proto.Request, port string){
	conn, err := grpc.NewClient(port)
		if err != nil{
			return
		}
		client := proto.NewConsensusClient(conn)
		ctx := context.Background()
		reply, err := client.RequestAccess(ctx, request)
		if reply.Access == false{
			log.Println("Something terrible happened!")
		}
}

func checkAvablePort(ports []string) (listener string, err error){
	for _, port := range ports {
        listener, err := net.Listen("tcp", port)

		if err == nil {
			listener.Close()
			return port, nil
		}
    }

	return "", errors.New("No avalable port!")
}


const (
	RELEASED int 	= 0
	HELD 			= 1
	WANTED 			= 2
)
