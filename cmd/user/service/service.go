package service

import (
	"context"

	"github.com/PorcoGalliard/eCommerce-Microservice/cmd/user/repository"
	"github.com/PorcoGalliard/eCommerce-Microservice/models"
)

type UserService struct {
	UserRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{
		UserRepo: userRepo,
	}
}

func (svc *UserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user, err := svc.UserRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (svc *UserService) CreateNewUser(ctx context.Context, user *models.User) (int64, error) {
	userID, err := svc.UserRepo.CreateNewUser(ctx, user)
	if err != nil {
		return 0, nil
	}

	return userID, nil
}