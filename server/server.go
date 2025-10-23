package main

import (
	proto "chitchat/grpc"
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
)

// this is the info the server remembers
type chitchatService struct {
	proto.UnimplementedChitchatServer
	time_stamp []int32 //slice of a timestamps
	next_user  int32
	user_feeds map[string]MessageFeed // List of user connections
	messages   []*proto.Message
}

type MessageFeed struct {
	stream proto.Chitchat_JoinServer
	done   chan bool
}

func main() {
	server := &chitchatService{
		time_stamp: []int32{},
		next_user:  0,
		user_feeds: make(map[string]MessageFeed),
		messages:   []*proto.Message{},
	} //creates server -> timestamp/users is a empty slice of int32

	server.messages = append(server.messages, &proto.Message{Text: "Test message"})

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

func (s *chitchatService) Join(in *proto.User, stream proto.Chitchat_JoinServer) error {
	s.time_stamp = append(s.time_stamp, 0)
	in.TimeStamp = s.time_stamp

	feed := MessageFeed{stream: stream, done: make(chan bool)}
	s.user_feeds[in.Username] = feed
	fmt.Println("User joining: " + in.Username)
	for _, message := range s.messages {
		stream.Send(message)
	}
	for {
		select {
		case <-feed.done:
			log.Println("Closing stream for: " + in.Username)
		case <-stream.Context().Done():
			log.Println("Disconnecting: " + in.Username)
			return nil
		}
	}
	//return &proto.User{TimeStamp: s.time_stamp, UserID: s.users[len(s.users)-1]}, nil
}

func (s *chitchatService) Leave(ctx context.Context, in *proto.User) (*proto.Empty, error) {

	// handle client's leave request and return emprty response
	for i, val := range s.time_stamp {
		if val < in.TimeStamp[i] {
			s.time_stamp[i] = in.TimeStamp[i] //checkin the timestamp stuff
		}
	}

	delete(s.user_feeds, in.Username)
	return &proto.Empty{}, nil
}

func (s *chitchatService) PostMessage(ctx context.Context, in *proto.Message) (*proto.Empty, error) {
	for i, val := range s.time_stamp {
		if i >= len(in.TimeStamp) {
			break
		}
		if val < in.TimeStamp[i] {
			s.time_stamp[i] = in.TimeStamp[i] //checkin the timestamp stuff
		}
	}

	in.TimeStamp = s.time_stamp
	s.messages = append(s.messages, in)
	for _, feed := range s.user_feeds {
		feed.stream.Send(in)
	}

	return &proto.Empty{}, nil
}
