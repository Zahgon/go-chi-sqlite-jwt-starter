package server

import (
	auth_handlers "go-chi-sqlite-jwt-starter/internal/auth/handlers"

	"github.com/gin-gonic/gin"
)

func authRouter(r *gin.RouterGroup) {
	r.POST("/register", gin.WrapF(auth_handlers.RegisterAccount))
	r.POST("/login", gin.WrapF(auth_handlers.Login))
}
