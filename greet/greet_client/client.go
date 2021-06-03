package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/simplesteph/grpc-go-course/greet/greetpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func main() {
	fmt.Printf("Hello  I a client\n")
	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure()) //connection to server

	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}

	defer cc.Close() //defer connection closure

	//Create Greet Client
	c := greetpb.NewGreetServiceClient(cc)

	//Make a unary request/response RPC call from client to server
	// doUnary(c)

	//Server Streaming RPC call
	//doServerStreaming(c)

	//Client streaming
	//doClientStreaming(c)

	//Bidirectional streaaming
	//doBiDiStreaming(c)

	//Unary withDeadline
	// doUnaryWithDeadline(c, 5) // Should complete
	doUnaryWithDeadline(c, 1) // Should not complete

}

func doUnary(c greetpb.GreetServiceClient) {
	req := &greetpb.GreetRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Nouru",
			LastName:  "Muneza",
		},
	}

	//Get response.
	res, err := c.Greet(context.Background(), req)

	if err != nil {
		log.Fatal("Error while calling greet RPC: ", err)
	}

	log.Printf("Response from Greet: %v", res.Result)
}

func doServerStreaming(c greetpb.GreetServiceClient) {
	fmt.Printf("Starting to do server streaming RPC...\n")

	req := &greetpb.GreetManyTimesRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Nouru",
			LastName:  "Muneza",
		},
	}

	resStream, err := c.GreetManyTimes(context.Background(), req)

	if err != nil {
		log.Fatal("Error while calling GreetManyTimes RPC: ", err)
	}

	for {
		msg, err := resStream.Recv()
		if err == io.EOF {
			log.Printf("We've reached the end of stream ")
			break
		}

		if err != nil {
			log.Fatal("Error while reading stream: ", err)
		}

		log.Printf("Response from GreetManyTimes: %v", msg.GetResult())
	}

}

func doClientStreaming(c greetpb.GreetServiceClient) {
	fmt.Printf("Starting to do client streaming RPC...\n")

	//Array of requests
	requests := []*greetpb.LongGreetRequest{
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Nouru",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "john",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Bob",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Turing",
			},
		},
	}

	clientStream, err := c.LongGreet(context.Background())
	if err != nil {
		log.Fatalf("Error while calling longGreet: %v", err)
	}

	// Iterate over the requests and send them requests to the server
	for _, req := range requests {
		clientStream.Send(req)
		time.Sleep(1000 * time.Millisecond)
	}

	//After sending all the requests, we close and receive a value from the server.
	res, err := clientStream.CloseAndRecv()
	if err != nil {
		log.Fatal("Error while receiving LongGreetResponse: ", err)
	}

	strings.Replace(res.Result, `\n`, "\n", -1)
	fmt.Printf("LongGreet Response: %v", res.Result)
}

func doBiDiStreaming(c greetpb.GreetServiceClient) {
	fmt.Printf("Starting to do biderectional streaming RPC...\n\n")

	//Array of requests
	requests := []*greetpb.GreetEveryoneRequest{
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Nouru",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "john",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Bob",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Turing",
			},
		},
	}

	// we create a stream by invoking the client
	stream, err := c.GreetEveryone(context.Background())
	if err != nil {
		log.Fatal("Error while creating stream: ", err)
		return
	}

	//A channel of structs
	waitChannel := make(chan struct{})

	// we send a bunch of messages to the server(go routine)
	go func() {
		for _, req := range requests {
			fmt.Printf("\nSending messages: %v\n", req)
			stream.Send(req)
			time.Sleep(1000 * time.Millisecond)
		}

		//After sending all the messages, we close the stream
		stream.CloseSend()
	}()

	//we receive messages from the server(go routine)
	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				//if EOF then close the waitChannel
				break
			}
			if err != nil {
				log.Fatal("Error while receiving : ", err)
				break
			}

			fmt.Printf("Received: %v\n", res.GetResult())
		}
		//Close channel if either last value is received or
		close(waitChannel)

	}()

	//block main thread till all go-routines finish executing
	<-waitChannel
}

func doUnaryWithDeadline(c greetpb.GreetServiceClient, timeOut int) {
	fmt.Printf("Starting to do unary with Deadline RPC ...\n\n")
	req := &greetpb.GreetWithDeadlineRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Nouru",
		},
	}

	//We are willing to wait for 5 secs for our request to be responded to
	context, cancel := context.WithTimeout(context.Background(), time.Duration(timeOut)*time.Second)
	defer cancel()

	res, err := c.GreetWithDeadline(context, req)
	if err != nil {
		statusErr, ok := status.FromError(err)
		if ok {
			if statusErr.Code() == codes.DeadlineExceeded {
				fmt.Printf("Timeout was hit! Dealine was exceeded")
			} else {
				fmt.Printf("unexpected error: %v", err)
			}
		} else {
			log.Fatal("Error while calling GreetWithDealine RPC: ", err)
		}
		return
	}
	log.Printf("Response from GreetWithDealine: %v", res.Result)
}
