package main

import (
	proto "chitchat/grpc"
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"google.golang.org/grpc"
)

// this is the info the server remembers
type chitchatService struct {
	proto.UnimplementedChitchatServer
	time_stamp []int32 //slice of a timestamps
	next_user  int32
	user_feeds map[int32]MessageFeed // List of user connections
	messages   []*proto.Message
}

type MessageFeed struct {
	stream proto.Chitchat_JoinServer
	done   chan bool
}

func main() {
	f, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	log.SetOutput(f)
	server := &chitchatService{
		time_stamp: []int32{},
		next_user:  0,
		user_feeds: make(map[int32]MessageFeed),
		messages:   []*proto.Message{},
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
	log.Println("Server started on " + listener.Addr().String())
	go s.shutdown_logger(grpc_server)
	err = grpc_server.Serve(listener)

	if err != nil {
		log.Fatal(err)
	}
}

func (s *chitchatService) shutdown_logger(grpc_server *grpc.Server) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop
	log.Printf("Server stopped with logical time stamp %v\n", s.time_stamp_string())
	log.Println("---------------------------------------------------------------")
	grpc_server.GracefulStop()
}

func (s *chitchatService) Join(in *proto.User, stream proto.Chitchat_JoinServer) error {
	if _, ok := s.user_feeds[in.Id]; ok {
		return fmt.Errorf("user id '%d' already taken", in.Id)
	}

	feed := MessageFeed{stream: stream, done: make(chan bool)}
	s.user_feeds[in.Id] = feed
	log.Printf("User joining: %d with logical time stamp %v\n", in.Id, s.time_stamp_string())

	join_message := &proto.Message{UserId: in.Id, Text: fmt.Sprintf("Participant %d joined Chit Chat at logical time %v", in.Id, s.time_stamp_string()), StatusUpdate: true}
	var receivers string
	for id, feed := range s.user_feeds {
		receivers = fmt.Sprintf("%v%d ", receivers, id)
		feed.stream.Send(join_message)
	}
	log.Printf("Sent join message of user %d to receivers [ %v] with logical time stamp %v\n", in.Id, receivers, s.time_stamp_string())

	for {
		select {
		case <-feed.done:
			log.Printf("Closing stream for %d with logical time stamp %v\n", in.Id, s.time_stamp_string())
		case <-feed.stream.Context().Done():
			// Just incase the user left by force closing
			delete(s.user_feeds, in.Id)

			log.Printf("Disconnecting %d with logical time stamp %v\n", in.Id, s.time_stamp_string())
			leave_message := &proto.Message{TimeStamp: s.time_stamp, UserId: in.Id, Text: fmt.Sprintf("Participant %d left Chit Chat at logical time %v", in.Id, s.time_stamp_string()), StatusUpdate: true}
			var receivers string
			for id, feed := range s.user_feeds {
				receivers = fmt.Sprintf("%v%d ", receivers, id)
				feed.stream.Send(leave_message)
			}
			log.Printf("Sent disconnect message of user %d to receivers [ %v] with logical time stamp %v\n", in.Id, receivers, s.time_stamp_string())
			return nil
		}
	}
}

// Lets the user know it's id number and initial time stamp, as this couldn't be sent by Join
func (s *chitchatService) GetUserID(ctx context.Context, in *proto.Empty) (*proto.User, error) {
	s.time_stamp = append(s.time_stamp, 0)
	user := &proto.User{Id: s.next_user, TimeStamp: s.time_stamp}
	s.next_user += 1
	return user, nil
}

func (s *chitchatService) Leave(ctx context.Context, in *proto.User) (*proto.Empty, error) {

	// handle client's leave request and return emprty response
	s.update_timestamp(in.TimeStamp)

	s.user_feeds[in.Id].done <- true
	delete(s.user_feeds, in.Id)
	return &proto.Empty{}, nil
}

func (s *chitchatService) PostMessage(ctx context.Context, in *proto.Message) (*proto.Empty, error) {
	s.update_timestamp(in.TimeStamp)

	if len(in.Text) > 128 {
		log.Printf("Received message exceeding 128 characters from %d with logical time stamp %v\n", in.UserId, s.time_stamp_string())
		return &proto.Empty{}, errors.New("message exceeds 128 characters")
	}
	log.Printf("Received message '%v' from %d with logical time stamp %v\n", in.Text, in.UserId, s.time_stamp_string())

	in.StatusUpdate = false
	s.messages = append(s.messages, in)
	var receivers string
	for id, feed := range s.user_feeds {
		receivers = fmt.Sprintf("%v%d ", receivers, id)
		feed.stream.Send(in)
	}
	log.Printf("Sent message from sender %d to receivers [ %v] with logical time stamp %v\n", in.UserId, receivers, s.time_stamp_string())

	return &proto.Empty{}, nil
}

func (s *chitchatService) update_timestamp(time_stamp_in []int32) {
	for i, val := range s.time_stamp {
		if i >= len(time_stamp_in) {
			break
		}
		if val < time_stamp_in[i] {
			s.time_stamp[i] = time_stamp_in[i]
		}
	}

	log.Printf("Updated time stamp: %v\n", s.time_stamp_string())
}

func (s *chitchatService) time_stamp_string() string {
	var time_stamp_str string
	for _, v := range s.time_stamp {
		time_stamp_str = fmt.Sprintf("%v%d ", time_stamp_str, v)
	}
	return "[ " + time_stamp_str + "]"
}
