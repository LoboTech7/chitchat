package main

import (
	"bufio"
	proto "chitchat/grpc"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var user *proto.User

func main() {
	// Create a connection to the grpc server, hosted on port 8080
	conn, err := grpc.NewClient("localhost:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}

	// New<service>Client. Is Database due to service name in proto file
	client := proto.NewChitchatClient(conn)

	// GetStudents rpc is defined in the proto file and implemented in the server.go file
	user_in, err := client.GetUserID(context.Background(), &proto.Empty{})
	if err != nil {
		log.Fatal(err)
	}
	user = user_in

	stream, err := client.Join(context.Background(), user)
	if err != nil {
		log.Fatal(err)
	}

	user.TimeStamp[user.Id] += 1

	reader := bufio.NewReader(os.Stdin)
	go get_input(reader, client)
	for {
		message, err := stream.Recv()
		if err != nil {
			log.Fatal(err)
		}
		user.TimeStamp[user.Id] += 1
		if !message.StatusUpdate {
			fmt.Printf("%d: '%v' with timestamp %v\n", message.UserId, message.Text, time_stamp_string(message.TimeStamp))
		} else {
			fmt.Println(message.Text)
		}
		update_timestamp(message.TimeStamp)
	}
}

func get_input(reader *bufio.Reader, client proto.ChitchatClient) {
	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		user.TimeStamp[user.Id] += 1
		input = strings.TrimSpace(input)
		fmt.Println(strings.Compare(input, ".quit"))
		if strings.Compare(input, ".quit") == 0 {
			user.TimeStamp[user.Id] += 1
			client.Leave(context.Background(), user)
			os.Exit(0)
			return
		}
		fmt.Println("Posting: " + input)
		user.TimeStamp[user.Id] += 1
		_, err = client.PostMessage(context.Background(), &proto.Message{Text: input, UserId: user.Id, TimeStamp: user.TimeStamp})
		if err != nil {
			log.Println(err)
		}
	}
}

func update_timestamp(time_stamp_in []int32) {
	for i, val := range time_stamp_in {
		if i >= len(user.TimeStamp) {
			user.TimeStamp = append(user.TimeStamp, val)
		}
		if val > user.TimeStamp[i] {
			user.TimeStamp[i] = val
		}
	}

	fmt.Printf("Updated time stamp: %v\n", time_stamp_string(user.TimeStamp))
}

func time_stamp_string(time_stamp_in []int32) string {
	var time_stamp_str string
	for _, v := range time_stamp_in {
		time_stamp_str = fmt.Sprintf("%v%d ", time_stamp_str, v)
	}
	return "[ " + time_stamp_str + "]"
}
