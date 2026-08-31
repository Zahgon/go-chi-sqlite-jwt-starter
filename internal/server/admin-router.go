package server

import (
	"go-chi-sqlite-jwt-starter/internal/auth"
	"go-chi-sqlite-jwt-starter/internal/models"
	"go-chi-sqlite-jwt-starter/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func adminRouter(r *gin.RouterGroup) {
	auth.UseAuthMiddleware(r)
	r.Use(adminOnly)

	r.GET("/test-token", gin.WrapF(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("You are successfully authenticated as an admin!"))
	}))
}

func adminOnly(c *gin.Context) {
	user := utils.GetUserFromContext(c.Writer, c.Request.Context())
	if user.Role != models.Admin {
		http.Error(c.Writer, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		c.Abort()
		return
	}
	c.Next()
}
