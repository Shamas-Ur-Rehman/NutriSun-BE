package middleware

import (
	"net/http"

	"Shamas/nutrisun/models"
	"Shamas/nutrisun/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RefreshPermissionCache is intentionally a no-op for now.
func RefreshPermissionCache(_ *gorm.DB) error {
	return nil
}

func RolePermissionMiddleware(moduleName, action string, masterDB *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authContext, ok := utils.GetCurrentUserContext(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
				Status:  http.StatusUnauthorized,
				Message: "User context not found",
			})
			c.Abort()
			return
		}

		userIDValue, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
				Status:  http.StatusUnauthorized,
				Message: "User not authenticated",
			})
			c.Abort()
			return
		}

		userID, userIDOK := userIDValue.(uint)
		if !userIDOK {
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Status:  http.StatusInternalServerError,
				Message: "Invalid user ID format",
			})
			c.Abort()
			return
		}

		var user models.User
		if err := masterDB.Select("id, role_id, is_active").First(&user, userID).Error; err != nil {
			c.JSON(http.StatusNotFound, utils.ErrorResponse{
				Status:  http.StatusNotFound,
				Message: "User not found",
			})
			c.Abort()
			return
		}

		if !user.IsActive {
			c.JSON(http.StatusForbidden, utils.ErrorResponse{
				Status:  http.StatusForbidden,
				Message: "User account is inactive",
			})
			c.Abort()
			return
		}

		var module models.Module
		if err := masterDB.Select("id").Where("business_id = ? AND branch_id = ? AND name = ?", authContext.BusinessID, authContext.BranchID, moduleName).First(&module).Error; err != nil {
			c.JSON(http.StatusForbidden, utils.ErrorResponse{
				Status:  http.StatusForbidden,
				Message: "Module not found",
			})
			c.Abort()
			return
		}

		var count int64
		if err := masterDB.Model(&models.Permission{}).
			Where("business_id = ? AND branch_id = ? AND role_id = ? AND module_id = ? AND action = ?", authContext.BusinessID, authContext.BranchID, user.RoleID, module.ID, action).
			Count(&count).Error; err != nil {
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Status:  http.StatusInternalServerError,
				Message: "Failed to check permission",
			})
			c.Abort()
			return
		}

		if count == 0 {
			c.JSON(http.StatusForbidden, utils.ErrorResponse{
				Status:  http.StatusForbidden,
				Message: "Insufficient permissions for this action",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
