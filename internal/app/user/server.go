package user

import (
	"context"
	"log"
	"net"

	db "github.com/Asuha-a/ProgrammingCourseMarket/internal/pkg/db/user"
	jwt "github.com/Asuha-a/ProgrammingCourseMarket/internal/pkg/jwt"
	pb "github.com/Asuha-a/ProgrammingCourseMarket/internal/pkg/pb/user"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
)

const (
	port = ":50051"
)

type server struct {
	pb.UnimplementedUserServer
}

func (s *server) CreateUser(ctx context.Context, in *pb.CreateUserRequest) (*pb.CreateUserReply, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(in.GetPassword()), bcrypt.MinCost)
	if err != nil {
		panic(err)
	}

	user := db.User{UUID: uuid.NewV4(), EMAIL: string(in.GetEmail()), PASSWORD: string(hash), PERMISSION: "normal"}
	result := db.DB.Create(&user)
	if result.Error != nil {
		log.Printf("failed to create user: %v", result.Error)
		return &pb.CreateUserReply{Token: ""}, result.Error
	}

	ss, err := jwt.CreateJWT(user)
	if err != nil {
		panic(err)
	}

	return &pb.CreateUserReply{Token: ss}, nil
}

func RunServer() {
	log.Println("test")
	db.Init()
	defer db.Close()

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterUserServer(s, &server{})
	log.Println("user grpc server running")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
