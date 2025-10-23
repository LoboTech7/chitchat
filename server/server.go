package main

import proto "chitchat/grpc"

type chitchatService struct {
	proto.UnimplementedChitchatServer
	time_stamp []int32 //slice of a timestamp!?
	user_id    int32
}

func main() {

}
