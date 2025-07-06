package grpc

import (
	// golang package
	"context"
	"log"
	"paymentfc/proto/userpb"
	"time"

	// external package
	"google.golang.org/grpc"
)

type UserClient interface {
	// GetUserInfoByUserID get user info by user id by given userID.
	//
	// It returns pointer of userpb.GetUserInfoResult, and nil error when successful.
	// Otherwise, nil pointer of userpb.GetUserInfoResult, and error will be returned.
	GetUserInfoByUserID(ctx context.Context, userID int64) (*userpb.GetUserInfoResult, error)
}

type userClient struct {
	Client userpb.UserServiceClient
}

// NewUserClient new user client.
//
// It returns pointer of UserClient when successful.
// Otherwise, nil pointer of UserClient will be returned.
func NewUserClient() UserClient {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed init grpc user client: %v", err)
	}

	client := userpb.NewUserServiceClient(conn)

	return &userClient{
		Client: client,
	}
}

// GetUserInfoByUserID get user info by user id by given userID.
//
// It returns pointer of userpb.GetUserInfoResult, and nil error when successful.
// Otherwise, nil pointer of userpb.GetUserInfoResult, and error will be returned.
func (uc userClient) GetUserInfoByUserID(ctx context.Context, userID int64) (*userpb.GetUserInfoResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	userInfo, err := uc.Client.GetUserInfoByUserID(ctx, &userpb.GetUserInfoRequest{
		UserId: userID,
	})
	if err != nil {
		return nil, err
	}

	return userInfo, nil
}