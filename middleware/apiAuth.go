package middleware

import (
	"fmt"
	"net/http"

	"Shamas/nutrisun/config"
	"Shamas/nutrisun/models"
	"Shamas/nutrisun/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func APIKeyMiddleware(masterDB *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
				Status:  http.StatusUnauthorized,
				Message: "API key is required",
			})
			c.Abort()
			return
		}
		var subscription models.Subscription
		if err := masterDB.Where("api_key = ?", apiKey).First(&subscription).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
					Status:  http.StatusUnauthorized,
					Message: "Invalid API key",
				})
				c.Abort()
				return
			}
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Status:  http.StatusInternalServerError,
				Message: "Failed to validate API key",
			})
			c.Abort()
			return
		}

		businessID := subscription.BusinessId
		branchID := subscription.BranchId
		var business models.Business
		if err := masterDB.Select("id", "db", "name_en").First(&business, businessID).Error; err != nil {
			c.JSON(http.StatusNotFound, utils.ErrorResponse{
				Status:  http.StatusNotFound,
				Message: "Business not found for this subscription",
			})
			c.Abort()
			return
		}
		dbName := business.DB
		if dbName == "" {
			dbName = utils.GetDefaultTenantDBName()
		}

		tenantDB := config.ConnectTenantSQLDB(dbName)
		if tenantDB == nil {
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Status:  http.StatusInternalServerError,
				Message: "Failed to connect tenant database",
			})
			c.Abort()
			return
		}
		employeeID := fmt.Sprintf("API-BUS-%d", businessID)
		employeeName := fmt.Sprintf("API: %s", business.NameEn)

		virtualUser := &models.User{
			ID:         0,
			FullName:   employeeName,
			EmployeeId: employeeID,
			BusinessId: businessID,
			BranchId:   branchID,
			RoleID:     0,
			IsActive:   true,
		}
		c.Set("business_id", businessID)
		c.Set("branch_id", branchID)
		c.Set("tenant_db", dbName)
		c.Set("tenant_db_conn", tenantDB)
		c.Set("business", business)
		c.Set("user", virtualUser)
		c.Set("subscription", subscription)
		c.Set("is_api_request", true)
		c.Set("user_id", fmt.Sprintf("api-%d", businessID))

		c.Next()
	}
}

func IsAPIRequest(c *gin.Context) bool {
	isAPI, exists := c.Get("is_api_request")
	if !exists {
		return false
	}
	return isAPI.(bool)
}
func GetSubscription(c *gin.Context) *models.Subscription {
	subscription, exists := c.Get("subscription")
	if !exists {
		return nil
	}
	return subscription.(*models.Subscription)
}
