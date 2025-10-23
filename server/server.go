package main

import proto "chitchat/grpc"

// this is the info the server remembers
type chitchatService struct {
	proto.UnimplementedChitchatServer
	time_stamp []int32 //slice of a timestamps
	users      []int32 // amount of users
}

func main() {
	server := &chitchatService{time_stamp: []int32{}, users: []int32{}} //creates server -> timestamp/users is a empty slice of int32

}
