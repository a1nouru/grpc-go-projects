package main

import (
	"context"
	"fmt"
	"go-grpc-course/calculator/calculatorpb"
	"io"
	"log"
	"math"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type server struct{}

// This is where server functionality is implemented. This function is called by the client.
func (*server) Sum(ctx context.Context, req *calculatorpb.SumRequest) (*calculatorpb.SumResponse, error) {
	fmt.Printf("Rreceived Sum RPC: %v", req)
	firstNumber := req.GetFirstNumber()
	secondNumber := req.GetSecondNunmber()

	sum := firstNumber + secondNumber

	res := &calculatorpb.SumResponse{
		Sum_Result: sum,
	}

	return res, nil
}

func (*server) PrimeNumberDecomposition(req *calculatorpb.PrimeNumberDecompositionRequest,
	stream calculatorpb.CalculatorService_PrimeNumberDecompositionServer) error {

	number := req.GetNumber()
	divisor := int64(2)

	for number > 1 {
		if number%divisor == 0 {
			stream.Send(&calculatorpb.PrimeNumberDecompositionResponse{
				Primefactor: divisor,
			})
			time.Sleep(1000 * time.Millisecond) //simulate latency in streaming

			number = number / divisor
		} else {
			divisor++
			fmt.Printf("Divisor has increased to %v\n", divisor)
		}
	}

	return nil
}

func (*server) ComputerAverage(stream calculatorpb.CalculatorService_ComputerAverageServer) error {
	fmt.Printf("Compute Average was invoked with a client stream request.....")

	var total, count int32

	for {
		req, err := stream.Recv()

		if err == io.EOF {
			//Once client envokes stream.CloseAndRecv(), we'll hit EOF and then return the response and close the stream
			return stream.SendAndClose(&calculatorpb.ComputerAverageResponse{
				Average: float64(total) / float64(count),
			})
		}

		if err != nil {
			log.Fatal("Error while reading client stream: ", err)
		}

		total += req.Number
		count++
	}
}

func (*server) FindMaximum(stream calculatorpb.CalculatorService_FindMaximumServer) error {
	fmt.Printf("Find maximum was invoked with a client stream request.....")
	maximum := int32(0)

	for {
		req, err := stream.Recv()

		if err == io.EOF {
			return nil
		}

		if err != nil {
			log.Fatal("Error while receiving req from client: ", err)
			return err
		}

		number := req.GetNumber()
		if number > maximum {
			maximum = number
			sendErr := stream.Send(&calculatorpb.FindMaximumResponse{
				Maximum: maximum,
			})
			if sendErr != nil {
				log.Fatal("Error while Sending maximum number : ", err)
				return err
			}
		}
	}
}

func FindMaximumNumber(numbers []int32) int32 {
	if numbers == nil {
		fmt.Printf("FindMaximumNumber: numbers array is empty")
		return -1
	}
	max := numbers[0]
	for _, number := range numbers {
		if number >= max {
			max = number
		}
	}

	return max
}

func (*server) SquareRoot(ctx context.Context, req *calculatorpb.SquareRootRequest) (
	*calculatorpb.SquareRootResponse, error) {
	fmt.Printf("Receieved SquareRoot RPC")

	number := req.GetNumber()
	if number < 0 {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Received a negative number: %v", number),
		)
	}

	return &calculatorpb.SquareRootResponse{
		NumberRoot: math.Sqrt(float64(number)),
	}, nil
}

func main() {
	fmt.Println("Calculator Server")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")

	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	//Create a GRPC server
	s := grpc.NewServer()
	calculatorpb.RegisterCalculatorServiceServer(s, &server{})

	// Register reflection service on gRPC server.
	reflection.Register(s)

	// Here we are binding the port to the grpc server
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to server: %v", err)
	}
}
