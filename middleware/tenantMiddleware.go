package middleware

import (
	"net/http"

	"Shamas/nutrisun/config"
	"Shamas/nutrisun/models"
	"Shamas/nutrisun/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func TenantMiddleware(masterDB *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDValue, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
				Status:  http.StatusUnauthorized,
				Message: "User not found in token",
			})
			c.Abort()
			return
		}

		userID, ok := userIDValue.(uint)
		if !ok {
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Status:  http.StatusInternalServerError,
				Message: "Invalid user ID format",
			})
			c.Abort()
			return
		}
		var user models.User
		if err := masterDB.First(&user, userID).Error; err != nil {
			c.JSON(http.StatusNotFound, utils.ErrorResponse{
				Status:  http.StatusNotFound,
				Message: "User not found",
			})
			c.Abort()
			return
		}
		var business models.Business
		if err := masterDB.Where("id = ?", user.BusinessId).First(&business).Error; err != nil {
			c.JSON(http.StatusNotFound, utils.ErrorResponse{
				Status:  http.StatusNotFound,
				Message: "Business not found",
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

		c.Set("business_id", user.BusinessId)
		c.Set("branch_id", user.BranchId)
		c.Set("tenant_db", dbName)
		c.Set("tenant_db_conn", tenantDB)
		c.Set("business", business)
		c.Set("user", &user)

		c.Next()
	}
}

func GetTenantDB(c *gin.Context) string {
	dbName, exists := c.Get("tenant_db")
	if !exists {
		return utils.GetDefaultTenantDBName()
	}
	return dbName.(string)
}

func GetTenantSQLDB(c *gin.Context) *gorm.DB {
	tenantDB, exists := c.Get("tenant_db_conn")
	if !exists {
		return nil
	}
	return tenantDB.(*gorm.DB)
}

func GetUser(c *gin.Context) *models.User {
	user, exists := c.Get("user")
	if !exists {
		return nil
	}
	return user.(*models.User)
}
func GetBusiness(c *gin.Context) *models.Business {
	business, exists := c.Get("business")
	if !exists {
		return nil
	}
	return business.(*models.Business)
}
