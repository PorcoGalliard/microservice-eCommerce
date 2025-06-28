package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/PorcoGalliard/eCommerce-Microservice/app/user/service"
	"github.com/PorcoGalliard/eCommerce-Microservice/infrastructure/log"
	"github.com/PorcoGalliard/eCommerce-Microservice/models"
	"github.com/PorcoGalliard/eCommerce-Microservice/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

type UserUsecase struct {
	UserService service.UserService
	JWTSecret string
}

func NewUserUsecase(userService *service.UserService, JWTSecret string) *UserUsecase {
	return &UserUsecase{
		UserService: *userService,
		JWTSecret: JWTSecret,
	}
}

func (uc *UserUsecase) GetUserByEmail (ctx context.Context, email string) (*models.User, error) {
	user, err := uc.UserService.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (uc *UserUsecase) GetUserByID (ctx context.Context, userID int64) (*models.User, error) {
	user, err := uc.UserService.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (uc *UserUsecase) RegisterUser (ctx context.Context, user *models.User) error {
	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"email": user.Email,
		}).Errorf("utils.HashPassword got an error at %v", err)
		return err
	}

	user.Password = hashedPassword
	_, err = uc.UserService.CreateNewUser(ctx, user)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"email": user.Name,
			"password": user.Password,
		}).Errorf("uc.UserService.CreateNewUser got an error at %v", err)
		return err
	}

	return nil
}

func (uc *UserUsecase) LoginUser (ctx context.Context, params *models.LoginParameter) (string, error) {
	user, err := uc.UserService.GetUserByEmail(ctx, params.Email)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"email": params.Email,
		}).Errorf("uc.UserService.GetUserByEmail got an error at %v", err)
	}

	isMatch, err := utils.CheckPasswordHash(user.Password, params.Password)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"email": params.Email,
		}).Errorf("utils.CheckPasswordHash got an error at %v", err)
		return "", err
	}

	if !isMatch {
		return "", errors.New("Invalid password")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp": time.Now().Add(time.Hour * 1).Unix(),
	})

	tokenString, err := token.SignedString([]byte(uc.JWTSecret))
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"email": params.Email,
		}).Errorf("token.SignedString got an error at %v", err)
	}

	return tokenString, nil
}

