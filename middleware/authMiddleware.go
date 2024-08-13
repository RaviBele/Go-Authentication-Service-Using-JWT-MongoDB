package middleware

import (
	"go-jwt-auth/helpers"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken := c.Request.Header.Get("token")
		if accessToken == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing token"})
			c.Abort()
			return
		}

		_, err := helpers.ValidateToken(accessToken)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			c.Abort()
			return
		}
	}
}
