package server

import (
	"context"
	"net/http"

	"go-chi-sqlite-jwt-starter/internal/auth"
	category_handlers "go-chi-sqlite-jwt-starter/internal/category/handlers"
	models "go-chi-sqlite-jwt-starter/internal/models"
	provider "go-chi-sqlite-jwt-starter/internal/provider"
	"go-chi-sqlite-jwt-starter/internal/utils"
	"go-chi-sqlite-jwt-starter/internal/validation"

	"github.com/gin-gonic/gin"
)

func categoryRouter(r *gin.RouterGroup) {
	auth.UseAuthMiddleware(r)

	r.GET("/list", gin.WrapF(category_handlers.ListCategories))
	r.POST("/create", gin.WrapF(category_handlers.CreateCategory))

	sub := r.Group("/:categoryID")
	sub.Use(CategoryCtx)
	sub.GET("", gin.WrapF(category_handlers.GetCategory))
	sub.PUT("", gin.WrapF(category_handlers.UpdateCategory))
	sub.DELETE("", gin.WrapF(category_handlers.DeleteCategory))
}

func CategoryCtx(c *gin.Context) {
	w := c.Writer
	categoryID := c.Param("categoryID")
	id, err := utils.StringToInt64(categoryID)
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		c.Abort()
		return
	}

	catgory, err := provider.Provider.CategoryService.GetCategory(id)
	if err != nil {
		http.Error(w, http.StatusText(404), http.StatusNotFound)
		c.Abort()
		return
	}

	user := utils.GetUserFromContext(w, c.Request.Context())
	err = validation.HasAccessToCategoryGroup(w, catgory.CategoryGroupID, user.ID)
	if err != nil {
		c.Abort()
		return
	}

	ctx := context.WithValue(c.Request.Context(), models.ContextKeys.Category, catgory)
	ctx = context.WithValue(ctx, models.ContextKeys.CategoryID, categoryID)
	c.Request = c.Request.WithContext(ctx)
	c.Next()
}
