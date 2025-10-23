package main

import (
	proto "chitchat/grpc"
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
)

// this is the info the server remembers
type chitchatService struct {
	proto.UnimplementedChitchatServer
	time_stamp []int32 //slice of a timestamps
	users      []int32 // amount of users
	messages   []string
}

func main() {
	server := &chitchatService{
		time_stamp: []int32{},
		users:      []int32{},
		messages:   []string{},
	} //creates server -> timestamp/users is a empty slice of int32
	server.start_server()
}

func (s *chitchatService) start_server() {
	grpc_server := grpc.NewServer()
	listener, err := net.Listen("tcp", ":8080")

	if err != nil {
		log.Fatal(err)
	}

	proto.RegisterChitchatServer(grpc_server, s)
	err = grpc_server.Serve(listener)

	if err != nil {
		log.Fatal(err)
	}
}

func (s *chitchatService) Join(ctx context.Context, in *proto.Empty) (*proto.User, error) {
	s.time_stamp = append(s.time_stamp, 0)
	s.users = append(s.users, int32(len(s.users)))
	return &proto.User{TimeStamp: s.time_stamp, UserId: s.users[len(s.users)-1]}, nil
}
