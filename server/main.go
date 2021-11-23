package main

import (
	"context"
	"fmt"

	//"io"
	"log"
	//"math"
	"errors"
	"net"
	"os"
	"strconv"
	"sync"

	pb "github.com/Miniim98/MiniProject3/proto"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedAuctionServer
}

type timestamp struct {
	time int32
	mu   sync.Mutex
}

type highestBid struct {
	amount int32
	id     int32
	mu     sync.Mutex
}

var Time timestamp
var listenaddr string
var id int
var bid highestBid
var length int
var ongoing bool

func (c *timestamp) UpTimestamp() {
	c.mu.Lock()
	defer c.mu.Unlock()
	Time.time++
	if Time.time == int32(length) {
		ongoing = false
	}
}

func (b *highestBid) WriteHighestBid(newId int32, newBid int32) {
	if !ongoing {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.amount = newBid
	b.id = newId
	Time.UpTimestamp()
}

func main() {
	Time.time = 0
	args := os.Args[1:]
	if len(args) < 3 {
		fmt.Println("Arguments required: <id> <listening address> <auction length>")
		os.Exit(1)
	}

	//args in order
	tempid, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Println("id must be an integer")
		os.Exit(1)
	}
	id = tempid

	templength, err := strconv.Atoi(args[2])
	if err != nil {
		fmt.Println("auction length must be an integer")
		os.Exit(1)
	}
	length = templength

	listenaddr := args[1]
	SetUpLog()

	lis, err := net.Listen("tcp", listenaddr)
	if err != nil {
		log.Fatalf("failed to listen on port : %v", err)
	}
	//use grpc to create a new server:
	grpcServer := grpc.NewServer()

	//register the server to be the one we have just created
	pb.RegisterAuctionServer(grpcServer, &server{})

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve gRPC server over port: %v", err)
	}

}

func (s *server) Bid(ctx context.Context, in *pb.BidRequest) (*pb.BidResponse, error) {
	var result string
	var error error

	//possible "extension": cheking if previous bid was made by same bidder
	if bid.amount <= in.Amount {
		result = "failure"
		error = errors.New("This bid is too low")
	} else {
		bid.WriteHighestBid(in.Id, in.Amount)
	}

	return &pb.BidResponse{Result: result, Timestamp: Time.time}, error
}

func (s *server) Result(ctx context.Context, in *pb.ResultRequest) (*pb.ResultResponse, error) {
	return &pb.ResultResponse{Ongoing: ongoing, Result: bid.amount,
		Timestamp: Time.time, BidderId: bid.id}, nil
}

func SetUpLog() {
	var filename = "serverLog" + strconv.Itoa(id)
	LOG_FILE := filename
	logFile, err := os.OpenFile(LOG_FILE, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Panic(err)
	}
	log.SetOutput(logFile)
}
