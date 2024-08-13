package helpers

import (
	"errors"
	"github.com/gin-gonic/gin"
)

func MatchUserTypeToUserID(c *gin.Context, userID string) error {
	userType := c.GetString("user_type")
	uID := c.GetString("uid")

	if userType == "USER" && uID != userID {
		return errors.New("Unauthorized user type")
	}

	return nil
}