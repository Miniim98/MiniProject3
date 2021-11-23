package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	pb "github.com/Miniim98/MiniProject3/proto"
	"google.golang.org/grpc"
)

var connections []pb.AuctionClient

var id int32
var currentBid int32

func main() {
	ports := [3]string{":5000", ":5001", ":5002"}
	tempid, err := strconv.Atoi(os.Args[1])

	if err != nil || tempid == 0 {
		fmt.Println("id must be an integer different from 0")
		os.Exit(1)
	}
	id = int32(tempid)
	//Dialing the server
	for _, port := range ports {
		conn, err := grpc.Dial(port, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			log.Fatalf("could not connect: %s", err)
		}
		defer conn.Close()

		connection := pb.NewAuctionClient(conn)
		connections = append(connections, connection)
	}
	for {
		HandleRequests()
	}
}

func HandleRequests() {
	fmt.Println("Write BID to place a bid and RESULT for results")
	var chosen string
	fmt.Scanln(&chosen)
	if chosen == "BID" {
		HandleBid()
	} else if chosen == "RESULT" {
		HandleResult()
	} else {
		fmt.Println("Please write BID or RESULT")
	}
}

func HandleBid() {
	fmt.Println("Please enter your bid:")
	var read string
	fmt.Scanln(&read)
	if read == "EXIT" {
		return
	}
	bid, err := strconv.Atoi(read)
	if err != nil {
		fmt.Println("That was not a number!")
		fmt.Println("Write EXIT to go back")
		HandleBid()
		return
	}
	if int32(bid) < currentBid {
		fmt.Println("Please enter a higher bid than your last!")
		fmt.Println("Write EXIT to go back")
		HandleBid()
		return
	} else {
		SendBidRequest(bid)
		currentBid = int32(bid)
	}

}

func HandleResult() {
	SendResultRequest()
}

func SendBidRequest(amount int) {
	for _, connection := range connections {
		response, err := connection.Bid(context.Background(), &pb.BidRequest{Amount: int32(amount), Id: id})
		if err != nil {
			log.Printf("Trouble sending bid request: %v", err)
			return
		}

		if response.Result != "succes" {
			log.Println(response.String())
			HandleBid()
		}
	}

}

func SendResultRequest() (int32, bool, int32) {
	var amount, winnerId, version int32
	version = 0
	var ongoing bool

	for _, connection := range connections {
		response, err := connection.Result(context.Background(), &pb.ResultRequest{})
		if err != nil {
			log.Printf("Trouble sending result request: %v /n", err)
		} else {
			if response.Timestamp > version {
				version = response.Timestamp
				amount = response.Result
				winnerId = response.BidderId
				ongoing = response.Ongoing
			}

		}
	}
	return amount, ongoing, winnerId
}

func SetUpLog() {
	var filename = "clientLog" + strconv.Itoa(int(id))
	LOG_FILE := filename
	logFile, err := os.OpenFile(LOG_FILE, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Panic(err)
	}
	log.SetOutput(logFile)
}
