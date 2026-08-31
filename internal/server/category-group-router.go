package server

import (
	"context"
	"net/http"

	"go-chi-sqlite-jwt-starter/internal/auth"
	category_group_handlers "go-chi-sqlite-jwt-starter/internal/category-group/handlers"
	models "go-chi-sqlite-jwt-starter/internal/models"
	provider "go-chi-sqlite-jwt-starter/internal/provider"
	"go-chi-sqlite-jwt-starter/internal/utils"

	"github.com/gin-gonic/gin"
)

func categoryGroupRouter(r *gin.RouterGroup) {
	auth.UseAuthMiddleware(r)

	r.GET("/list", gin.WrapF(category_group_handlers.ListCategoryGroups))
	r.POST("/create", gin.WrapF(category_group_handlers.CreateCategoryGroup))

	sub := r.Group("/:categoryGroupID")
	sub.Use(CategoryGroupCtx)
	sub.GET("", gin.WrapF(category_group_handlers.GetCategoryGroup))
	sub.POST("/rename", gin.WrapF(category_group_handlers.RenameCategoryGroup))
	sub.DELETE("", gin.WrapF(category_group_handlers.DeleteCategoryGroup))
}

func CategoryGroupCtx(c *gin.Context) {
	w := c.Writer
	categoryGroupID := c.Param("categoryGroupID")
	id, err := utils.StringToInt64(categoryGroupID)
	if err != nil {
		http.Error(w, "Invalid category group ID", http.StatusBadRequest)
		c.Abort()
		return
	}

	user := utils.GetUserFromContext(w, c.Request.Context())
	catgoryGroup, err := provider.Provider.CategoryGroupService.GetCategoryGroupForUser(id, user.ID)
	if err != nil {
		http.Error(w, http.StatusText(404), http.StatusNotFound)
		c.Abort()
		return
	}

	ctx := context.WithValue(c.Request.Context(), models.ContextKeys.CategoryGroup, catgoryGroup)
	ctx = context.WithValue(ctx, models.ContextKeys.CategoryGroupID, categoryGroupID)
	c.Request = c.Request.WithContext(ctx)
	c.Next()
}
