package handlers

import (
	"auth-service/internal/logger"
	"auth-service/internal/models"
	"auth-service/internal/services"
	"auth-service/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SignupHandler(c *gin.Context) {
	logger.Log.Info("signup endpoint hit")
	var user models.User

	if err := c.ShouldBindJSON(&user); err != nil {
		logger.Log.Error(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	err := services.Signup(user)

	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}
	logger.Log.Info(
		"user signup successful",
	)

	c.JSON(http.StatusCreated, gin.H{
		"message": "user created",
	})
}

func LoginHandler(c *gin.Context) {

	var request models.User

	if err := c.ShouldBindJSON(&request); err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	user, err := services.Login(
		request.Email,
		request.Password,
	)

	if err != nil {

		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})

		return
	}

	token, err := utils.GenerateJWT(user.Email)

	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to generate token",
		})

		return
	}
	logger.Log.Info(
		"user login successful",
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "login successful",
		"token":   token,
		"user":    user.Email,
	})
}
