package routes

import (
	"go-jwt-auth/controllers"

	"github.com/gin-gonic/gin"
)

func AuthRoutes(router *gin.Engine) {
	router.POST("users/signup", controllers.SignUp())
	router.POST("users/login", controllers.Login())
}
