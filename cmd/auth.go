package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Aergiaaa/rollet/internal/database"
	"github.com/Aergiaaa/rollet/internal/env"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
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

type googleProfile struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
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
	tokenStr, err := token.SignedString([]byte(app.jwtSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, loginResponse{
		Token:  tokenStr,
		UserID: existingUser.Id,
	})
}

func (app *app) loginGoogle(c *gin.Context) {
	// pick google env
	clientID := env.GetEnvString("GOOGLE_OAUTH_CLIENT_ID", "")
	clientSecret := env.GetEnvString("GOOGLE_OAUTH_CLIENT_SECRET", "")
	redirectURL := env.GetEnvString("GOOGLE_OAUTH_REDIRECT_URL", "")

	if clientID == "" || clientSecret == "" || redirectURL == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Google OAuth environment variables not set"})
		return
	}

	cfg := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email",
		},
		Endpoint: google.Endpoint,
	}

	code := c.Query("code")
	state := c.Query("state")

	if code == "" {
		authURL := cfg.AuthCodeURL(state, oauth2.AccessTypeOffline)
		c.Redirect(http.StatusFound, authURL)
		return
	}

	ctx := context.Background()
	googleToken, err := cfg.Exchange(ctx, code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token"})
		return
	}

	client := cfg.Client(ctx, googleToken)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil || resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
		return
	}
	defer resp.Body.Close()

	var p googleProfile
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode user info"})
		return
	}

	user, err := app.models.Users.GetByGoogleID(p.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		return
	}

	if user == nil {
		newUser := &database.User{
			GoogleId: p.ID,
			Name:     p.Name,
			Password: "",
		}
		err = app.models.Users.Insert(newUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": user.Id,
		"expr":   time.Now().Add(time.Hour * 3).Unix(),
	})

	// Sign the token with the secret
	tokenStr, err := token.SignedString([]byte(app.jwtSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, loginResponse{
		Token:  tokenStr,
		UserID: user.Id,
	})
}
