package middleware

import (
	"net/http"
	"strings"

	"Shamas/nutrisun/config"
	"Shamas/nutrisun/models"
	"Shamas/nutrisun/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CustomerJWTAuth(masterDB *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
				Status:  http.StatusUnauthorized,
				Message: "Authorization header is required",
			})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
				Status:  http.StatusUnauthorized,
				Message: "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		claims, err := utils.ValidateJWT(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
				Status:  http.StatusUnauthorized,
				Message: "Invalid token: " + err.Error(),
			})
			c.Abort()
			return
		}

		if claims.Service != "customer" {
			c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
				Status:  http.StatusUnauthorized,
				Message: "Invalid customer token",
			})
			c.Abort()
			return
		}

		var business models.Business
		if err := masterDB.Select("id", "db", "name_en", "name_ar").First(&business, claims.BusinessID).Error; err != nil {
			c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
				Status:  http.StatusUnauthorized,
				Message: "Business not found",
			})
			c.Abort()
			return
		}

		dbName := claims.DbName
		if strings.TrimSpace(dbName) == "" {
			dbName = business.DB
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

		var customer models.Customer
		if err := tenantDB.Where("id = ? AND business_id = ? AND branch_id = ? AND status = ?", claims.UserID, claims.BusinessID, claims.BranchID, "active").First(&customer).Error; err != nil {
			c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
				Status:  http.StatusUnauthorized,
				Message: "Customer not found or inactive",
			})
			c.Abort()
			return
		}

		c.Set("customer_id", customer.ID)
		c.Set("customer", &customer)
		c.Set("business_id", customer.BusinessId)
		c.Set("branch_id", customer.BranchId)
		c.Set("tenant_db", dbName)
		c.Set("tenant_db_conn", tenantDB)
		c.Set("business", business)
		c.Next()
	}
}

func GetCustomer(c *gin.Context) *models.Customer {
	customer, exists := c.Get("customer")
	if !exists {
		return nil
	}
	return customer.(*models.Customer)
}
