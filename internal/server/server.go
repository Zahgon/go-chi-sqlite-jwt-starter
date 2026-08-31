package server

import (
	"go-chi-sqlite-jwt-starter/internal/auth"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Initialize() *gin.Engine {
	log.Println("Initializing server...")
	defer log.Println("Server initialized")

	r := gin.New()
	useGlobalMiddleware(r)
	auth.InitializeTokenVerifier()

	categoryRouter(r.Group("/category"))
	categoryGroupRouter(r.Group("/category-group"))
	adminRouter(r.Group("/admin"))
	authRouter(r.Group("/auth"))

	return r
}

func useGlobalMiddleware(r *gin.Engine) {
	r.Use(requestID)
	r.Use(realIP)
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, ".")
	})
	r.HEAD("/health", func(c *gin.Context) {
		c.String(http.StatusOK, ".")
	})
}
