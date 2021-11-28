package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"errors"

	pb "github.com/Miniim98/MiniProject3/proto"
	"google.golang.org/grpc"
)

var connections []pb.AuctionClient

var id int32
var currentBid int32
var writeResponses, readResponses = 2, 2


func main() {
	ports := [3]string{":5000", ":5001", ":5002"}
	tempid, err := strconv.Atoi(os.Args[1])

	if err != nil || tempid == 0 {
		fmt.Println("id must be an integer different from 0")
		os.Exit(1)
	}
	id = int32(tempid)
	SetUpLog()
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
		if (SendBidRequest(bid)) {
			fmt.Printf("Your bid of %d has been accepted", bid)
			fmt.Println()
			currentBid = int32(bid)
			HandleRequests()
		} else {
			fmt.Printf("Your bid of %d was not accepted", bid)
			fmt.Println()
			HandleBid()
		}
	}

}

func HandleResult() {
	amount, ongoing, winner, err := SendResultRequest()

	if err != nil {
		fmt.Printf("Something went wrong.  please try again")
		return;
	} else if (ongoing) {
		fmt.Printf("The auction is ongoing. The current bid is %d", amount)
		fmt.Println()
	} else {
		fmt.Printf("The auction is over. Winning bet was %d, made by %d", amount, winner)
		fmt.Println()
	}
}

//this method should also check for versioning via timestamps
func SendBidRequest(amount int) (bool) {
	var successes = 0
	var version int32
	version = 0
	for _, connection := range connections {
		response, err := connection.Bid(context.Background(), &pb.BidRequest{Amount: int32(amount), Id: id})
		if err != nil {
			log.Printf("Trouble sending bid request: %v", err)
		} else if response.Timestamp > version {
			successes = 0
			version = response.Timestamp
			if response.Result {
				successes++	
			} else {
				log.Println(response.String())
			}
		} else if response.Timestamp == version {
			if response.Result {
				successes++
			} else {
				log.Println(response.String())
			}
		}
	}
	
	if successes >= writeResponses {
		return true
	} else {
		log.Println("An insufficient amount of responses were received. It is not possible to write to the servers")
		log.Printf("I received %d responses", successes)
		return false
	}
}

func SendResultRequest() (int32, bool, int32, error) {
	var amount, winnerId, version int32
	var answers = 0
	version = -1
	var ongoing bool

	for _, connection := range connections {
		response, err := connection.Result(context.Background(), &pb.ResultRequest{})
		if err != nil {
			log.Printf("Trouble sending result request: %v /n", err)
		} else {
			if response.Timestamp > version {
				answers = 1
				version = response.Timestamp
				amount = response.Result
				winnerId = response.BidderId
				ongoing = response.Ongoing
			} else if response.Timestamp == version {
				answers++
			}
		}
	}
	if answers >= readResponses {
		return amount, ongoing, winnerId, nil
	} else {
		log.Println("an insuffucuient amount of responses was received. it is not possible to read from the servers")
		log.Printf("I received %d responses", answers)
		log.Println()
		var err error
		err = errors.New("not enough responsesor")
		return 0, false, 0, err
	}
	
}



func SetUpLog() {
	var filename = "clientLog" + strconv.Itoa(int(id)) + ".txt"
	LOG_FILE := filename
	logFile, err := os.OpenFile(LOG_FILE, os.O_APPEND|os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Panic(err)
	}
	log.SetOutput(logFile)
}
