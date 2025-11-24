package main

import (
	"net/http"
	"time"

	"github.com/Aergiaaa/rollet/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

type registerRequest struct {
	GoogleID string `json:"google_id"`
	Name     string `json:"name" binding:"min=3"`
	Password string `json:"password" binding:"min=8"`
}

type loginRequest struct {
	GoogleID string `json:"google_id"`
	Name     string `json:"name" binding:"min=3"`
	Password string `json:"password" binding:"min=8"`
}

type loginResponse struct {
	Token  string `json:"token"`
	UserID int    `json:"user_id"`
}

func (app *app) register(c *gin.Context) {
	var req registerRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
	}

	req.Password = string(hashPassword)
	user := database.User{
		GoogleId: req.GoogleID,
		Name:     req.Name,
		Password: req.Password,
	}

	err = app.models.Users.Insert(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register user"})
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfuly",
		"user":    user,
	})
}

func (app *app) login(c *gin.Context) {
	var req loginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	existingUser, err := app.models.Users.GetByGoogleID(req.GoogleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
	}

	if existingUser == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Google ID or password"})
	}

	err = bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Google ID or password"})
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": existingUser.Id,
		"expr":   time.Now().Add(time.Hour * 3).Unix(),
	})

	tokenString, err := token.SignedString([]byte(app.jwtSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
	}

	c.JSON(http.StatusOK, loginResponse{
		Token:  tokenString,
		UserID: existingUser.Id,
	})
}
