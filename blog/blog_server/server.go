package main

import (
	"context"
	"fmt"
	"go-grpc-course/blog/blogpb"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var collection *mongo.Collection

type server struct{}

type blogItem struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	AuthorID string             `bson:"author_id"`
	Content  string             `bson:"content"`
	Title    string             `bson:"title"`
}

func mapDataToBlog(data *blogItem) *blogpb.Blog {

	return &blogpb.Blog{
		Id:       data.ID.Hex(),
		AuthorId: data.AuthorID,
		Content:  data.Content,
		Title:    data.Title,
	}
}

func (*server) CreateBlog(ctx context.Context, req *blogpb.CreateBlogRequest) (*blogpb.CreateBlogResponse, error) {
	fmt.Printf("CreateBlog called by client....\n")
	blog := req.GetBlog()

	data := blogItem{
		AuthorID: blog.GetAuthorId(),
		Title:    blog.GetTitle(),
		Content:  blog.GetContent(),
	}

	res, err := collection.InsertOne(context.Background(), data)

	if err != nil {
		return nil, status.Errorf(
			codes.Internal, fmt.Sprintf("Internal error: %v", err),
		)
	}

	objectId, ok := res.InsertedID.(primitive.ObjectID)
	fmt.Printf("OID: %v", objectId)
	if !ok {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Cannot convert to ObjectId"),
		)
	}

	return &blogpb.CreateBlogResponse{
		Blog: &blogpb.Blog{
			Id:       objectId.Hex(),
			AuthorId: blog.GetAuthorId(),
			Content:  blog.GetContent(),
		},
	}, nil
}

func (*server) ReadBlog(ctx context.Context, req *blogpb.ReadBlogRequest) (*blogpb.ReadBlogResponse, error) {
	fmt.Printf("ReadBlog called by client....\n")

	blogId := req.GetBlogId()
	oid, err := primitive.ObjectIDFromHex(blogId)
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprint("Cannot parse objectID"),
		)
	}

	//create an empty struct
	data := &blogItem{}

	res := collection.FindOne(context.Background(), bson.M{"_id": oid})
	if err := res.Decode(data); err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Cannot find blog with specified ID: %v", err),
		)
	}

	return &blogpb.ReadBlogResponse{
		Blog: &blogpb.Blog{
			Id:       data.ID.Hex(),
			AuthorId: data.AuthorID,
			Title:    data.Title,
			Content:  data.Content,
		},
	}, nil
}

func (*server) UpdateBlog(ctx context.Context, req *blogpb.UpdateBlogRequest) (*blogpb.UpdateBlogResponse, error) {
	fmt.Printf("UpdateBlog called by client...\n")
	blog := req.GetBlog()

	oid, err := primitive.ObjectIDFromHex(blog.GetId())
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Cannot parse ID for blog %v ", blog),
		)
	}

	//create and empty struct
	data := &blogItem{}

	res := collection.FindOne(context.Background(), bson.M{"_id": oid})
	if err := res.Decode(data); err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Cannot find blog with specified ID: %v", err),
		)
	}

	//Update stored values
	data.AuthorID = blog.GetAuthorId()
	data.Content = blog.GetContent()
	data.Title = blog.GetTitle()

	_, updatedErr := collection.ReplaceOne(context.Background(), bson.M{"_id": oid}, data)
	if updatedErr != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Cannot update object in MongoDB: %v", updatedErr),
		)
	}

	return &blogpb.UpdateBlogResponse{
		Blog: mapDataToBlog(data),
	}, nil
}

func (*server) DeleteBlog(ctx context.Context, req *blogpb.DeleteBlogRequest) (*blogpb.DeleteBlogResponse, error) {
	fmt.Printf("DeleteBlog called ...\n\n")

	oid, err := primitive.ObjectIDFromHex(req.GetBlogId())
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Cannot parse ID: %v", oid.Hex()),
		)
	}

	deleteResult, deleteErr := collection.DeleteOne(context.Background(), bson.M{"_id": oid})
	if deleteErr != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Cannot delete object in MongoDB: %v", deleteErr),
		)
	}

	if deleteResult.DeletedCount == 0 {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("ObjectId to be deleted is not found: %v", oid.Hex()),
		)
	}

	return &blogpb.DeleteBlogResponse{
		BlogId: oid.Hex(),
	}, nil
}

func (*server) ListBlog(req *blogpb.ListBlogRequest, stream blogpb.BlogService_ListBlogServer) error {
	fmt.Printf("ListBlog called...\n")

	cur, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		return status.Errorf(
			codes.Internal,
			fmt.Sprintf("Unknown internal error: %v", err),
		)
	}
	defer cur.Close(context.Background())

	for cur.Next(context.Background()) {
		data := &blogItem{}

		if err := cur.Decode(data); err != nil {
			return status.Errorf(
				codes.Internal,
				fmt.Sprintf("Error while decoding data from MongoDB: %v", err),
			)
		}
		stream.Send(&blogpb.ListBlogResponse{Blog: mapDataToBlog(data)})
	}

	if err := cur.Err(); err != nil {
		return status.Errorf(
			codes.Internal,
			fmt.Sprintf("Unknown internal error: %v", err),
		)
	}

	return nil
}

func main() {
	fmt.Println("Blog Service ...")
	// If server crashes, we get the file name and line number in terminal
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Connect to mongoDB
	fmt.Println("Connecting to mongoDB")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal("Failed to conntect mongodb: ", err)
	}

	collection = client.Database("mydb").Collection("blog")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)

	}
	//Create a GRPC server
	opt := []grpc.ServerOption{}
	s := grpc.NewServer(opt...)
	blogpb.RegisterBlogServiceServer(s, &server{})

	go func() {
		fmt.Println("Starting server... ")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to server: %v", err)
		}
	}()

	// Wait for Control C to be triggered in order to stop the server.
	ch := make(chan os.Signal, 1)   // channel of type os.Signal and length 1.
	signal.Notify(ch, os.Interrupt) // notify channel ch with os.Interrupt

	// Block main thread untill signal is relayed to ch by signal.Notify
	<-ch

	fmt.Println("\nClosing mongodb connection...")
	if err = client.Disconnect(ctx); err != nil {
		panic(err)
	}

	fmt.Println("Stopping server...")
	s.Stop()

	fmt.Println("Closing the listener...")
	lis.Close()
}
