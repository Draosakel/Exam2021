package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	// this has to be the same as the go.mod module,
	// followed by the path to the folder the proto file is in.
	gRPC "github.com/Draosakel/Exam2021/proto"

	"google.golang.org/grpc"
)

type Server struct {
	gRPC.UnimplementedTemplateServer        // You need this line if you have a server
	name                             string // Not required but useful if you want to name your server
	port                             string // Not required but useful if your server needs to know what port it's listening to

	keyValueMap map[int64]int64 // value that clients can increment.
	mutex       sync.Mutex      // used to lock the server to avoid race conditions.
}

// flags are used to get arguments from the terminal. Flags take a value, a default value and a description of the flag.
// to use a flag then just add it as an argument when running the program.
var serverName = flag.String("name", "default", "Senders name") // set with "-name <name>" in terminal
var port = flag.String("port", "5400", "Server port")           // set with "-port <port>" in terminal

func main() {

	// f := setLog() //uncomment this line to log to a log.txt file instead of the console
	// defer f.Close()

	// This parses the flags and sets the correct/given corresponding values.
	flag.Parse()
	fmt.Println(".:server is starting:.")

	// starts a goroutine executing the launchServer method.
	launchServer()

	// code here is unreachable because launchServer occupies the current thread.
}

func launchServer() {
	log.Printf("Server %s: Attempts to create listener on port %s\n", *serverName, *port)

	// Create listener tcp on given port or default port 5400
	list, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", *port))
	if err != nil {
		log.Printf("Server %s: Failed to listen on port %s: %v", *serverName, *port, err) //If it fails to listen on the port, run launchServer method again with the next value/port in ports array
		return
	}

	// makes gRPC server using the options
	// you can add options here if you want or remove the options part entirely
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

	// makes a new server instance using the name and port from the flags.
	server := &Server{
		name:        *serverName,
		port:        *port,
		keyValueMap: make(map[int64]int64), // gives default value, but not sure if it is necessary
	}

	gRPC.RegisterTemplateServer(grpcServer, server) //Registers the server to the gRPC server.

	log.Printf("Server %s: Listening at %v\n", *serverName, list.Addr())

	if err := grpcServer.Serve(list); err != nil {
		log.Fatalf("failed to serve %v", err)
	}
	// code here is unreachable because grpcServer.Serve occupies the current thread.
}

func (s *Server) Get(ctx context.Context, GetMessage *gRPC.GetMessage) (*gRPC.Result, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	getKey := int64(GetMessage.GetKey())
	mapValue := s.keyValueMap[getKey]
	return &gRPC.Result{Value: mapValue}, nil
}

func (s *Server) Put(ctx context.Context, PutMessage *gRPC.PutMessage) (*gRPC.Ack, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	putKey := PutMessage.GetKey()
	putValue := PutMessage.GetValue()
	s.keyValueMap[putKey] = putValue

	return &gRPC.Ack{Success: true}, nil
}

// Get preferred outbound ip of this machine
// Usefull if you have to know which ip you should dial, in a client running on an other computer
func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

// sets the logger to use a log.txt file instead of the console
func setLog() *os.File {
	// Clears the log.txt file when a new server is started
	if err := os.Truncate("log.txt", 0); err != nil {
		log.Printf("Failed to truncate: %v", err)
	}

	// This connects to the log file/changes the output of the log informaiton to the log.txt file.
	f, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(f)
	return f
}
