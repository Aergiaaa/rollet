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
	Name     string `json:"name" binding:"required,min=3"`
	Password string `json:"password" binding:"min=8"`
}

type loginRequest struct {
	GoogleID string `json:"google_id"`
	Name     string `json:"name" binding:"required,min=3"`
	Password string `json:"password" binding:"min=8"`
}

type loginResponse struct {
	Token  string `json:"token"`
	UserID int    `json:"user_id"`
}

func (app *app) register(c *gin.Context) {
	// Bind and validate input
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hash the password
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
	}

	// Create user object
	user := database.User{
		GoogleId: req.GoogleID,
		Name:     req.Name,
		Password: string(hashPassword),
	}

	// Insert user into database
	err = app.models.Users.Insert(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register user"})
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfuly",
		"user":    user,
	})
}

func (app *app) loginDefault(c *gin.Context) {

	// Bind and validate input
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Retrieve user by Google ID
	existingUser, err := app.models.Users.GetByGoogleID(req.GoogleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		return
	}

	// Check if user exists
	if existingUser == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Google ID or password"})
		return
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Google ID or password"})
		return
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": existingUser.Id,
		"expr":   time.Now().Add(time.Hour * 3).Unix(),
	})

	// Sign the token with the secret
	tokenString, err := token.SignedString([]byte(app.jwtSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, loginResponse{
		Token:  tokenString,
		UserID: existingUser.Id,
	})
}

func (app *app) loginGoogle(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Google login not implemented yet"})
	return
}
