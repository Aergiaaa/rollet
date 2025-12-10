package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func (app *app) routes() http.Handler {
	g := gin.Default()

	config := cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	g.Use(cors.New(config))

	v1 := g.Group("/v1")
	{
		v1.POST("/random/default", app.createRandomize)
		v1.POST("/random/custom", app.createCustomRandomize)

		v1.POST("/auth/register", app.register)
		v1.POST("/auth/login/default", app.loginDefault)
		// TODO: Implement Google OAuth login
		v1.POST("/auth/google", app.googleAuth)

		v1.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status": "ok",
			})
		})
	}

	authGroup := v1.Group("/")
	authGroup.Use(app.AuthMiddleware())
	{
		authGroup.POST("/user/random", app.createRandomize)
		authGroup.GET("/user/history", app.getHistory)
	}

	{
		g.GET("/swagger/*any", func(ctx *gin.Context) {
			if ctx.Request.RequestURI == "/swagger/" {
				ctx.Redirect(http.StatusFound, "/swagger/index.html")
				return
			}
			ginSwagger.WrapHandler(swaggerFiles.Handler,
				ginSwagger.URL(fmt.Sprintf("http://%s:%d/swagger/doc.json",
					app.host, app.port)))(ctx)
		})
	}

	return g
}
