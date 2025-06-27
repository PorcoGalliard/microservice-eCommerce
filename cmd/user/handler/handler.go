package handler

import (
	"net/http"

	"github.com/PorcoGalliard/eCommerce-Microservice/cmd/user/usecase"
	"github.com/PorcoGalliard/eCommerce-Microservice/infrastructure/log"
	"github.com/PorcoGalliard/eCommerce-Microservice/models"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	UserUsecase usecase.UserUsecase
}

func NewUserHandler(userUsecase usecase.UserUsecase) *UserHandler {
	return &UserHandler{
		UserUsecase: userUsecase,
	}
}

func (h *UserHandler) Register(c *gin.Context) {
	var param models.RegisterParameter
	if err := c.ShouldBindJSON(&param); err != nil {
		log.Logger.Info(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error_message": err.Error(),
		})
		return
	}

	if len(param.Password) < 8 ||
	len(param.ConfirmPassword) < 8 {
		log.Logger.Info("Invalid password length")
		c.JSON(http.StatusBadRequest, gin.H{
			"error_message": "Password must more than 8 characters",
		})
		return
	}

	if param.Password != param.ConfirmPassword {
		log.Logger.Info("Invalid password equity confirmation")
		c.JSON(http.StatusBadRequest, gin.H{
			"error_message": "Password and Confirm Password Not Match",
		})
		return
	}

	user, err := h.UserUsecase.GetUserByEmail(c.Request.Context() , param.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error_message": err.Error(),
		})
		return
	}

	if user != nil && user.ID != 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error_message": "Email already exist",
		})
		return
	}

	err = h.UserUsecase.RegisterUser(c.Request.Context(), &models.User{
		Name: param.Name,
		Email: param.Email,
		Password: param.Password,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error_message": err.Error(),
		})
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User successfully registered",
	})
}

func (h *UserHandler) Login(c *gin.Context) {
	var params models.LoginParameter
	if err := c.ShouldBindJSON(&params); err != nil {
		log.Logger.Info(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error_message": "Invalid login input parameter",
		})
		return
	}

	if len(params.Password) < 8 {
		log.Logger.Info("Invalid password length")
		c.JSON(http.StatusBadRequest, gin.H{
			"error_message": "Invalid password length",
		})
		return
	}

	token, err := h.UserUsecase.LoginUser(c.Request.Context(), &params)
	if err != nil {
		log.Logger.Error(err.Error())
		c.JSON(http.StatusUnauthorized, gin.H{
			"error_message": "Email atau Password salah",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}

func (h *UserHandler) Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "OK",
	})
}