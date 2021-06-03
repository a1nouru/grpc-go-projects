package main

import (
	"context"
	"fmt"
	"go-grpc-course/blog/blogpb"
	"io"
	"log"

	"google.golang.org/grpc"
)

func main() {
	fmt.Printf("Hello  I am client\n\n")
	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure()) //connection to server

	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}

	defer cc.Close() //defer connection closure

	//Create Greet Client
	c := blogpb.NewBlogServiceClient(cc)

	//Create blog
	// createBlog(c)

	//Read blog
	// readBlog(c)

	//Update blog
	// updateBlog(c)

	//delete blog
	// deleteBlog(c)

	//List blog
	listBlog(c)
}

func createBlog(c blogpb.BlogServiceClient) {
	fmt.Printf("Create blog Client called\n")
	blog := &blogpb.Blog{
		AuthorId: "Nouru",
		Title:    "Learning about Blogs",
		Content:  "Creating our first blog",
	}

	res, err := c.CreateBlog(context.Background(), &blogpb.CreateBlogRequest{Blog: blog})
	if err != nil {
		log.Fatal("unexected error: ", err)
	}

	fmt.Printf("\nBlog created: %v", res.Blog)
}

func readBlog(c blogpb.BlogServiceClient) {
	fmt.Printf("Read blog client called\n")

	blogId := "60b4066e58ae45070601eb67"
	res, err := c.ReadBlog(context.Background(), &blogpb.ReadBlogRequest{BlogId: blogId})
	if err != nil {
		log.Fatal("Err while calling ReadBlog: ", err)
	}

	fmt.Printf("\nBlog with id: %v was read from db\n%v", blogId, res.Blog)
}

func updateBlog(c blogpb.BlogServiceClient) {
	fmt.Printf("Update blog client...\n")

	blog := &blogpb.Blog{
		AuthorId: "Umutesi",
		Title:    "Econ 101",
		Content:  "Learning the power of markets",
		Id:       "60b4066e58ae45070601eb67",
	}

	res, err := c.UpdateBlog(context.Background(), &blogpb.UpdateBlogRequest{Blog: blog})
	if err != nil {
		log.Fatal("Err while updating blog:\n", err)
	}

	fmt.Printf("Updated blog from %v\n\n to %v", blog, res.Blog)
}

func deleteBlog(c blogpb.BlogServiceClient) {
	fmt.Printf("DeleteBlog called by client...\n\n")
	blogId := "60b4066e58ae45070601eb67"

	res, err := c.DeleteBlog(context.Background(), &blogpb.DeleteBlogRequest{BlogId: blogId})
	if err != nil {
		log.Fatal("Err while deleting blog :\n", err)
	}

	fmt.Printf("blogId deleted: %v\n", res.GetBlogId())
}

func listBlog(c blogpb.BlogServiceClient) {
	fmt.Printf("List blog called... \n\n")

	resStream, err := c.ListBlog(context.Background(), &blogpb.ListBlogRequest{})
	if err != nil {
		log.Fatal("Err while calling ListBlog ", err)
	}

	for {
		res, err := resStream.Recv()
		if err == io.EOF {
			log.Printf("We've reached the end of stream ")
			break
		}
		if err != nil {
			log.Fatal("Error while reading stream: ", err)
		}

		log.Printf("Blog: %v\n\n", res.Blog)
	}
}
