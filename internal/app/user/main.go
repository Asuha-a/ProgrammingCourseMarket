package user

import (
	"context"
	"errors"
	"log"
	"net"

	db "github.com/Asuha-a/ProgrammingCourseMarket/internal/pkg/db/user"
	jwt "github.com/Asuha-a/ProgrammingCourseMarket/internal/pkg/jwt"
	pb "github.com/Asuha-a/ProgrammingCourseMarket/internal/pkg/pb/user"
	"github.com/gofrs/uuid"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
)

const (
	port = ":50051"
)

type server struct {
	pb.UnimplementedUserServer
}

func (s *server) ListUsers(rect *pb.ListUsersRequest, stream pb.User_ListUsersServer) error {
	log.Println("ListUsers running")
	var users []db.User
	result := db.DB.Find(&users)
	log.Println("got all users")
	log.Println(result.Error)
	if result.Error != nil {
		log.Fatalf("failed to list users: %v", result.Error)
		return result.Error
	}
	for _, user := range users {
		if err := stream.Send(&pb.ListUsersReply{
			Uuid:         user.UUID.String(),
			Name:         user.NAME,
			Introduction: user.INTRODUCTION,
			Email:        user.EMAIL,
			Permission:   user.PERMISSION,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *server) GetUser(ctx context.Context, in *pb.GetUserRequest) (*pb.GetUserReply, error) {
	var user db.User
	result := db.DB.First(&user, "UUID = ?", in.GetUuid())
	log.Println(user)
	if result.Error != nil {
		log.Printf("failed to get a user: %v", result.Error)
		return &pb.GetUserReply{}, result.Error
	}
	return &pb.GetUserReply{
		Uuid:         user.UUID.String(),
		Name:         user.NAME,
		Introduction: user.INTRODUCTION,
		Email:        user.EMAIL,
		Permission:   user.PERMISSION,
	}, nil
}

func (s *server) CreateUser(ctx context.Context, in *pb.CreateUserRequest) (*pb.CreateUserReply, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(in.GetPassword()), bcrypt.MinCost)
	if err != nil {
		panic(err)
	}

	user := db.User{
		UUID:         uuid.Must(uuid.NewV4()),
		NAME:         in.GetName(),
		INTRODUCTION: in.GetIntroduction(),
		EMAIL:        in.GetEmail(),
		PERMISSION:   "normal",
		PASSWORD:     string(hash),
	}
	log.Println(user)
	result := db.DB.Create(&user)
	if result.Error != nil {
		log.Printf("failed to create user: %v", result.Error)
		return &pb.CreateUserReply{Token: ""}, result.Error
	}

	ss, err := jwt.CreateJWT(user)
	if err != nil {
		panic(err)
	}

	return &pb.CreateUserReply{
		Token:        ss,
		Uuid:         user.UUID.String(),
		Name:         user.NAME,
		Introduction: user.INTRODUCTION,
		Email:        user.EMAIL,
		Permission:   user.PERMISSION,
	}, nil
}

func (s *server) UpdateUser(ctx context.Context, in *pb.UpdateUserRequest) (*pb.UpdateUserReply, error) {
	var user db.User
	uuid, _, err := jwt.ParseJWT(in.GetToken())
	if err != nil {
		log.Printf("failed to parse jwt: %v", err)
	}
	result := db.DB.First(&user, "UUID = ?", uuid)
	if result.Error != nil {
		log.Printf("failed to update user: %v", result.Error)
		return &pb.UpdateUserReply{Token: ""}, result.Error
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.GetPassword()), bcrypt.MinCost)
	if err != nil {
		panic(err)
	}
	user.NAME = in.GetName()
	user.INTRODUCTION = in.GetIntroduction()
	user.EMAIL = in.GetEmail()
	user.PASSWORD = string(hash)
	db.DB.Save(&user)

	ss, err := jwt.CreateJWT(user)
	if err != nil {
		panic(err)
	}

	return &pb.UpdateUserReply{
		Token:        ss,
		Uuid:         user.UUID.String(),
		Name:         user.NAME,
		Introduction: user.INTRODUCTION,
		Email:        user.EMAIL,
		Permission:   user.PERMISSION,
	}, nil
}

func (s *server) DeleteUser(ctx context.Context, in *pb.DeleteUserRequest) (*empty.Empty, error) {
	var user db.User
	uuid, _, err := jwt.ParseJWT(in.GetToken())
	if uuid.String() != in.GetUuid() {
		return new(empty.Empty), errors.New("invalid access")
	}
	if err != nil {

		return new(empty.Empty), err
	}
	result := db.DB.First(&user, "UUID = ?", uuid)
	if result.Error != nil {

		return new(empty.Empty), result.Error
	}
	result = db.DB.Delete(&user, "UUID = ?", uuid)
	log.Println(user, result.Error)
	if result.Error != nil {
		log.Printf("failed to delete a user: %v", result.Error)
		return new(empty.Empty), result.Error
	}
	return new(empty.Empty), nil
}

func RunServer() {

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
