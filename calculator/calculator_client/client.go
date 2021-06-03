package main

import (
	"context"
	"fmt"
	"go-grpc-course/calculator/calculatorpb"
	"io"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func main() {
	fmt.Printf("Hello  I a client")
	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure()) //connection to server

	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}

	defer cc.Close() //defer connection closure

	//Create Greet Client
	c := calculatorpb.NewCalculatorServiceClient(cc)

	//Make a unary request/response RPC call from client to server
	// doUnary(c)

	//Server streaming RPC
	// doServerStreaming(c)

	//Client streaming RPC
	// doClientStreaming(c)

	//Bidirectional streaaming
	// doBiDiStreaming(c)

	// Valid call
	dovalidSquareRootUnary(c, 2)

	//Invalid call
	doInvalidSquareUnary(c, -32)

}

func doUnary(c calculatorpb.CalculatorServiceClient) {

	req := &calculatorpb.SumRequest{
		FirstNumber:   23,
		SecondNunmber: 60,
	}

	//Call Sum method from server.go
	res, err := c.Sum(context.Background(), req)

	if err != nil {
		log.Fatal("Error while calling Sum RPC: ", err)
	}

	log.Printf("Response from Sum: %v", res.Sum_Result)
}

func doServerStreaming(c calculatorpb.CalculatorServiceClient) {

	req := &calculatorpb.PrimeNumberDecompositionRequest{
		Number: 5,
	}

	resStream, err := c.PrimeNumberDecomposition(context.Background(), req)

	if err != nil {
		log.Fatal("Error while calling PrimeNumberDecomposition RPC: ", err)
	}

	for {
		msg, err := resStream.Recv()
		if err == io.EOF {
			log.Printf("\n\nWe've reached the end of stream ")
			break
		}

		if err != nil {
			log.Fatal("\nError while reading stream: ", err)
		}

		log.Printf("Response from PrimeNumberDecomposition: %v", msg.Primefactor)
	}

}

func doClientStreaming(c calculatorpb.CalculatorServiceClient) {
	fmt.Printf("Starting to do client streaming RPC...\n")

	// Array of requests
	numbers := []*calculatorpb.ComputerAverageRequest{
		{
			Number: 1,
		},
		{
			Number: 2,
		},
		{
			Number: 3,
		},
		{
			Number: 4,
		},
	}

	clientStream, err := c.ComputerAverage(context.Background())
	if err != nil {
		log.Fatalf("Error while calling longGreet: %v", err)
	}

	// Iterate over the requests and send them requests to the server
	for _, req := range numbers {
		clientStream.Send(req)
		time.Sleep(1000 * time.Millisecond)
	}

	//After sending all the requests, we close and receive a value from the server.
	res, err := clientStream.CloseAndRecv()
	if err != nil {
		log.Fatal("Error while receiving LongGreetResponse: ", err)
	}

	fmt.Printf("Compute average Response: %v\n\n", res.Average)
}

func doBiDiStreaming(c calculatorpb.CalculatorServiceClient) {
	fmt.Printf("Starting to do client streaming RPC...\n\n")

	// Array of requests
	numbers := []*calculatorpb.FindMaximumRequest{
		{
			Number: 1,
		},
		{
			Number: 5,
		},
		{
			Number: 3,
		},
		{
			Number: 6,
		},
		{
			Number: 2,
		},
		{
			Number: 20,
		},
	}

	waitChannel := make(chan struct{})

	clientStream, err := c.FindMaximum(context.Background())
	if err != nil {
		log.Fatalf("Error while calling FindMaximumRequests: %v", err)
	}

	//Sending numbers to servers
	go func() {
		for _, number := range numbers {
			fmt.Printf("\nSending messages: %v\n", number)
			clientStream.Send(number)
		}

		//After sending all the messages, we close the stream which will alert EOF on the server side.
		clientStream.CloseSend()
	}()

	//Receiving maximum responses
	go func() {
		fmt.Printf("Responses: ")
		for {
			res, err := clientStream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal("Error while reciving maximum value from server: ", err)
				break
			}
			fmt.Print(res.Maximum, ",")
		}

		//Close channel to unblock the main thread after we've received all maximum values from server
		close(waitChannel)
	}()

	//Block the main thread until all other go routines have ended executing.
	<-waitChannel
}

func dovalidSquareRootUnary(c calculatorpb.CalculatorServiceClient, n int32) {
	fmt.Printf("Starting to do sqareRoot Unary RPC ...\n\n")

	//Correct call
	res, err := c.SquareRoot(context.Background(), &calculatorpb.SquareRootRequest{Number: n})

	if err != nil {
		log.Fatal("Error while recieving Response: ", err)
	}
	fmt.Printf("Square Root of %v is %v\n", n, res.GetNumberRoot())
}

func doInvalidSquareUnary(c calculatorpb.CalculatorServiceClient, n int32) {
	fmt.Printf("\n\nStarting to do sqareRoot with Invalid entry Unary RPC ...\n\n")

	res, err := c.SquareRoot(context.Background(), &calculatorpb.SquareRootRequest{Number: n})

	if err != nil {
		resErr, ok := status.FromError(err)
		if ok {
			//Actuall error from gRPC (user error)
			fmt.Printf(resErr.Message())
			fmt.Printf(resErr.Code().String())
			if resErr.Code() == codes.InvalidArgument {
				fmt.Printf("We sent a negative number!")
				return
			}
		} else {
			log.Fatal("Err while calling SquareRoot: ", resErr)
		}
	}
	fmt.Printf("Result of square root %v: %v\n", n, res.GetNumberRoot())
}
