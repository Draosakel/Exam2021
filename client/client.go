package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	gRPC "github.com/Draosakel/Exam2021/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Same principle as in client. Flags allows for user specific arguments/values
var clientsName = flag.String("name", "default", "Senders name")
var serverPort = flag.String("server", "5400", "Tcp server")

var server gRPC.TemplateClient  //the server
var ServerConn *grpc.ClientConn //the server connection

func main() {
	//parse flag/arguments
	flag.Parse()

	fmt.Println("--- CLIENT APP ---")

	//log to file instead of console
	//f := setLog()
	//defer f.Close()

	//connect to server and close the connection when program closes
	fmt.Println("--- join Server ---")
	ConnectToServer()
	defer ServerConn.Close()

	//start the biding
	parseInput()
}

// connect to server
func ConnectToServer() {

	//dial options
	//the server is not using TLS, so we use insecure credentials
	//(should be fine for local testing but not in the real world)
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithBlock(), grpc.WithTransportCredentials(insecure.NewCredentials()))

	//dial the server, with the flag "server", to get a connection to it
	log.Printf("client %s: Attempts to dial on port %s\n", *clientsName, *serverPort)
	conn, err := grpc.Dial(fmt.Sprintf(":%s", *serverPort), opts...)
	if err != nil {
		log.Printf("Fail to Dial : %v", err)
		return
	}

	// makes a client from the server connection and saves the connection
	// and prints rather or not the connection was is READY
	server = gRPC.NewTemplateClient(conn)
	ServerConn = conn
	log.Println("the connection is: ", conn.GetState().String())
}

func parseInput() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Type 'GET [key]' to get a value or 'PUT [key] [value] to add a new key/value to the map")
	fmt.Println("--------------------")

	//Infinite loop to listen for clients input.
	for {
		fmt.Print("-> ")

		//Read input into var input and any errors into err
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		input = strings.TrimSpace(input) //Trim input
		inputs := strings.Split(input, " ")
		if inputs[0] == "GET" {
			key, errKey := strconv.ParseInt(inputs[1], 10, 64)
			if errKey != nil || key < 1 {
				fmt.Println("Enter a valid key (Must be a natural number)")
				continue
			}
			get(key)
		} else if inputs[0] == "PUT" {
			key, errKey := strconv.ParseInt(inputs[1], 10, 64)
			if errKey != nil || key < 1 {
				fmt.Println("Enter a valid key (Must be a natural number)")
				continue
			}
			value, errVal := strconv.ParseInt(inputs[2], 10, 64)
			if errVal != nil || key < 1 {
				fmt.Println("Enter a valid value (Must be a natural number)")
				continue
			}
			put(key, value)
		}
	}
}

func get(key int64) {
	getMessage := &gRPC.GetMessage{
		Key: key,
	}
	value, err := server.Get(context.Background(), getMessage)
	if err != nil {
		fmt.Println("New error received")
		fmt.Println(err)
	}
	fmt.Printf("Key: %d returns Value: %d\n", key, value.Value)
}

func put(key int64, value int64) {
	putMessage := &gRPC.PutMessage{
		Key:   key,
		Value: value,
	}
	ack, err := server.Put(context.Background(), putMessage)
	if err != nil {
		fmt.Println("New error received")
		fmt.Println(err)
	}
	if ack.Success == true {
		fmt.Printf("Key/Value was succesfully added to the map\n")
	} else {
		fmt.Printf("Key/Value was NOT succesfully added to the map\n")
	}
}

// sets the logger to use a log.txt file instead of the console
func setLog() *os.File {
	f, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(f)
	return f
}
