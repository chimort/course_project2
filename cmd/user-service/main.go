package main

import (
    "context"
    "fmt"
    "log"
    "net"

    "github.com/chimort/course_project2/api/proto/sharedpb"
    "github.com/chimort/course_project2/api/proto/userpb"
    "google.golang.org/grpc"
)

type userServer struct {
    userpb.UnimplementedUserServiceServer
    users map[string]*sharedpb.User
}

func newUserServer() *userServer {
    return &userServer{
        users: make(map[string]*sharedpb.User),
    }
}

func (s *userServer) CreateUser(ctx context.Context, req *userpb.CreateUserRequest) (*userpb.CreateUserResponse, error) {
    user := req.GetUser()
    if _, exists := s.users[user.Id]; exists {
        return &userpb.CreateUserResponse{Response: "user already exists"}, nil
    }
    s.users[user.Id] = user
    log.Printf("✅ User created: %s (%s)", user.Username, user.Id)
    return &userpb.CreateUserResponse{Response: "user created successfully"}, nil
}

func main() {
    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("❌ failed to listen: %v", err)
    }

    grpcServer := grpc.NewServer()
    userpb.RegisterUserServiceServer(grpcServer, newUserServer())

    fmt.Println("🚀 UserService running on :50051")
    if err := grpcServer.Serve(lis); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}