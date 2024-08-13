package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"

	"go-jwt-auth/database"
	"go-jwt-auth/helpers"
	"go-jwt-auth/models"
)

var validate = validator.New()

type SignUPResponse struct {
	UserID       string `json:"user_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func HashPassword(password string) string {
	hashPassword, _ := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(hashPassword)
}

func VerifyPassword(userPassword, providedPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(userPassword), []byte(providedPassword))

	if err != nil {
		return false, err
	}

	return true, nil
}

func SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var user models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationError := validate.Struct(user)
		if validationError != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationError.Error()})
			return
		}

		password := HashPassword(*user.Password)
		user.Password = &password

		emailCount, err := database.UserCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		if err != nil {
			log.Printf("Failed to count documents: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		phoneCount, err := database.UserCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
		if err != nil {
			log.Printf("Failed to count documents: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if emailCount > 0 || phoneCount > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "email or phone already exists"})
			return
		}

		user.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.UserID = user.ID.Hex()

		token, refershToken, _ := helpers.GenerateAllTokens(*user.Email, *user.FirstName, *user.LastName, *user.UserType, user.UserID)

		user.RefreshToken = &refershToken
		_, err = database.UserCollection.InsertOne(ctx, user)
		if err != nil {
			msg := fmt.Sprintf("User not created: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		response := SignUPResponse{
			UserID:       user.UserID,
			AccessToken:  token,
			RefreshToken: *user.RefreshToken,
		}

		c.JSON(http.StatusOK, response)
	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var user models.User
		var foundUser models.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := database.UserCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "email or password are not valid"})
			return
		}

		isPasswordValid, msg := VerifyPassword(*foundUser.Password, *user.Password)

		if !isPasswordValid {
			log.Printf("Password verification failed: %s", msg.Error())
			c.JSON(http.StatusUnauthorized, gin.H{"error": "password is not valid"})
			return
		}

		token, refershToken, _ := helpers.GenerateAllTokens(*foundUser.Email, *foundUser.FirstName, *foundUser.LastName, *foundUser.UserType, foundUser.UserID)

		var updateObj primitive.D
		updateObj = append(updateObj, bson.E{"refresh_token", refershToken})
		updatedAt, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{"updated_at", updatedAt})
		filter := bson.M{"user_id": foundUser.UserID}
		upsert := true
		opt := &options.UpdateOptions{
			Upsert: &upsert,
		}

		_, err := database.UserCollection.UpdateOne(ctx, filter, bson.D{bson.E{"$set", updateObj}}, opt)
		if err != nil {
			msg := fmt.Sprintf("User not updated: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		response := SignUPResponse{
			UserID:       foundUser.UserID,
			AccessToken:  token,
			RefreshToken: refershToken,
		}

		c.JSON(http.StatusOK, response)
	}
}

func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := database.UserCollection.Find(ctx, bson.D{})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error aggregating pipeline"})
			return
		}

		var allUsers []bson.M

		if err = result.All(ctx, &allUsers); err != nil {
			log.Fatalf("Error aggregating pipeline %v", err.Error())
		}

		c.JSON(http.StatusOK, allUsers)
	}
}

func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("user_id")

		if err := helpers.MatchUserTypeToUserID(c, userID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User
		err := database.UserCollection.FindOne(ctx, bson.M{"userid": userID}).Decode(&user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, user)
	}
}
