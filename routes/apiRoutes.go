package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"Shamas/nutrisun/middleware"
)

func SetupAPIRoutes(r *gin.Engine, masterDB *gorm.DB) {
	api := r.Group("/api/v1")
	api.Use(middleware.APIKeyMiddleware(masterDB))
	api.Use(middleware.RateLimiterMiddleware())

	// Health check endpoint for API
	api.GET("/health", func(c *gin.Context) {
		user := middleware.GetUser(c)
		subscription := middleware.GetSubscription(c)
		c.JSON(200, gin.H{
			"status":          "ok",
			"system_user":     user.FullName,
			"business_id":     user.BusinessId,
			"branch_id":       user.BranchId,
			"subscription_id": subscription.ID,
		})
	})
}
