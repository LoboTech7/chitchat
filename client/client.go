package main

import (
	"bufio"
	proto "chitchat/grpc"
	"context"
	"fmt"
	"log"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Create a connection to the grpc server, hosted on port 8080
	conn, err := grpc.NewClient("localhost:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}

	// New<service>Client. Is Database due to service name in proto file
	client := proto.NewChitchatClient(conn)

	// GetStudents rpc is defined in the proto file and implemented in the server.go file
	stream, err := client.Join(context.Background(), &proto.User{Username: os.Args[1]})
	if err != nil {
		log.Fatal(err)
	}
	reader := bufio.NewReader(os.Stdin)
	go get_input(reader, client)
	for {
		message, err := stream.Recv()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(message.Text)
	}

}

func get_input(reader *bufio.Reader, client proto.ChitchatClient) {
	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Posting: " + input)
		client.PostMessage(context.Background(), &proto.Message{Text: input})
	}
}
