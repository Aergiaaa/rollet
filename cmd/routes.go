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
		g.POST("/random", app.createRandomize)
	}

	authGroup := v1.Group("/")
	authGroup.Use(app.AuthMiddleware())
	{

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
