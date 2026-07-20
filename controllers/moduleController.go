package controllers

import (
	"Shamas/nutrisun/dtos"
	"Shamas/nutrisun/models"
	"Shamas/nutrisun/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
)

func CreateModule(c *gin.Context, masterDB *gorm.DB) {
	var req dtos.ModuleReq
	authContext, ok := utils.GetCurrentUserContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Status:  http.StatusUnauthorized,
			Message: "User context not found",
		})
		return
	}
	businessID := authContext.BusinessID
	branchID := authContext.BranchID

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}
	var existingModule models.Module
	if err := masterDB.Where("business_id = ? AND branch_id = ? AND LOWER(name) = LOWER(?)", businessID, branchID, req.Name).First(&existingModule).Error; err == nil {
		c.JSON(http.StatusConflict, utils.ErrorResponse{
			Status:  http.StatusConflict,
			Message: "Module name already exists (case-insensitive)",
		})
		return
	}

	module := models.Module{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		BusinessId:  businessID,
		BranchId:    branchID,
	}

	if err := masterDB.Create(&module).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to create module: " + err.Error(),
		})
		return
	}

	moduleRes := dtos.ModuleRes{
		ID:          module.ID,
		Name:        module.Name,
		DisplayName: module.DisplayName,
		Description: module.Description,
		BusinessId:  module.BusinessId,
		BranchId:    module.BranchId,
		CreatedAt:   module.CreatedAt,
		UpdatedAt:   module.UpdatedAt,
	}

	c.JSON(http.StatusCreated, utils.SuccessResponse{
		Status:  http.StatusCreated,
		Message: "Module created successfully",
		Data: map[string]interface{}{
			"module": moduleRes,
		},
	})
}

func GetAllModules(c *gin.Context, masterDB *gorm.DB) {
	page := utils.StringToInt(c.DefaultQuery("page", "1"))
	perPage := utils.StringToInt(c.DefaultQuery("per_page", "10"))
	search := c.Query("search")
	authContext, ok := utils.GetCurrentUserContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Status:  http.StatusUnauthorized,
			Message: "User context not found",
		})
		return
	}
	businessID := authContext.BusinessID
	branchID := authContext.BranchID

	query := masterDB.Model(&models.Module{}).Where("business_id = ? AND branch_id = ?", businessID, branchID)

	if search != "" {
		query = query.Where("LOWER(name) LIKE LOWER(?) OR LOWER(display_name) LIKE LOWER(?)", "%"+search+"%", "%"+search+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to count modules",
		})
		return
	}

	offset := (page - 1) * perPage
	totalPages := int((total + int64(perPage) - 1) / int64(perPage))

	var modules []models.Module
	if err := query.Offset(offset).Limit(perPage).Order("created_at DESC").Find(&modules).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to retrieve modules",
		})
		return
	}

	moduleList := make([]dtos.ModuleRes, 0, len(modules))
	for _, module := range modules {
		moduleList = append(moduleList, dtos.ModuleRes{
			ID:          module.ID,
			Name:        module.Name,
			DisplayName: module.DisplayName,
			Description: module.Description,
			BusinessId:  module.BusinessId,
			BranchId:    module.BranchId,
			CreatedAt:   module.CreatedAt,
			UpdatedAt:   module.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Modules retrieved successfully",
		Data: map[string]interface{}{
			"modules": moduleList,
			"pagination": map[string]interface{}{
				"total_records": total,
				"total_pages":   totalPages,
				"page":          page,
				"per_page":      perPage,
			},
		},
	})
}

func GetModuleByID(c *gin.Context, masterDB *gorm.DB) {
	moduleID := c.Param("id")
	authContext, ok := utils.GetCurrentUserContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Status:  http.StatusUnauthorized,
			Message: "User context not found",
		})
		return
	}
	businessID := authContext.BusinessID
	branchID := authContext.BranchID

	var module models.Module
	if err := masterDB.Where("business_id = ? AND branch_id = ?", businessID, branchID).First(&module, moduleID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, utils.ErrorResponse{
				Status:  http.StatusNotFound,
				Message: "Module not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to retrieve module",
		})
		return
	}
	var permissions []models.Permission
	masterDB.Where("business_id = ? AND branch_id = ? AND module_id = ?", businessID, branchID, moduleID).Find(&permissions)

	permList := make([]dtos.PermissionRes, 0, len(permissions))
	for _, perm := range permissions {
		permList = append(permList, dtos.PermissionRes{
			ID:         perm.ID,
			ModuleID:   perm.ModuleID,
			Action:     perm.Action,
			BusinessId: perm.BusinessId,
			BranchId:   perm.BranchId,
			CreatedAt:  perm.CreatedAt,
			UpdatedAt:  perm.UpdatedAt,
		})
	}

	moduleRes := dtos.ModuleRes{
		ID:          module.ID,
		Name:        module.Name,
		DisplayName: module.DisplayName,
		Description: module.Description,
		BusinessId:  module.BusinessId,
		BranchId:    module.BranchId,
		CreatedAt:   module.CreatedAt,
		UpdatedAt:   module.UpdatedAt,
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Module retrieved successfully",
		Data: map[string]interface{}{
			"module":      moduleRes,
			"permissions": permList,
		},
	})
}

func UpdateModule(c *gin.Context, masterDB *gorm.DB) {
	moduleID := c.Param("id")
	authContext, ok := utils.GetCurrentUserContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Status:  http.StatusUnauthorized,
			Message: "User context not found",
		})
		return
	}
	businessID := authContext.BusinessID
	branchID := authContext.BranchID

	var req dtos.ModuleUpdateReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	var module models.Module
	if err := masterDB.Where("business_id = ? AND branch_id = ?", businessID, branchID).First(&module, moduleID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, utils.ErrorResponse{
				Status:  http.StatusNotFound,
				Message: "Module not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to retrieve module",
		})
		return
	}

	if req.DisplayName != "" {
		module.DisplayName = req.DisplayName
	}
	if req.Description != "" {
		module.Description = req.Description
	}

	if err := masterDB.Save(&module).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to update module: " + err.Error(),
		})
		return
	}

	moduleRes := dtos.ModuleRes{
		ID:          module.ID,
		Name:        module.Name,
		DisplayName: module.DisplayName,
		Description: module.Description,
		BusinessId:  module.BusinessId,
		BranchId:    module.BranchId,
		CreatedAt:   module.CreatedAt,
		UpdatedAt:   module.UpdatedAt,
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Module updated successfully",
		Data: map[string]interface{}{
			"module": moduleRes,
		},
	})
}
func DeleteModule(c *gin.Context, masterDB *gorm.DB) {
	moduleID := c.Param("id")
	authContext, ok := utils.GetCurrentUserContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Status:  http.StatusUnauthorized,
			Message: "User context not found",
		})
		return
	}
	businessID := authContext.BusinessID
	branchID := authContext.BranchID

	var module models.Module
	if err := masterDB.Where("business_id = ? AND branch_id = ?", businessID, branchID).First(&module, moduleID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, utils.ErrorResponse{
				Status:  http.StatusNotFound,
				Message: "Module not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to retrieve module",
		})
		return
	}
	if err := masterDB.Where("business_id = ? AND branch_id = ? AND module_id = ?", businessID, branchID, moduleID).Delete(&models.Permission{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to delete module permissions",
		})
		return
	}
	if err := masterDB.Delete(&module).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to delete module: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Module deleted successfully",
		Data:    nil,
	})
}
func GetModulePermissions(c *gin.Context, masterDB *gorm.DB) {
	moduleID := c.Param("id")
	authContext, ok := utils.GetCurrentUserContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Status:  http.StatusUnauthorized,
			Message: "User context not found",
		})
		return
	}
	businessID := authContext.BusinessID
	branchID := authContext.BranchID
	var module models.Module
	if err := masterDB.Select("id, name").Where("business_id = ? AND branch_id = ?", businessID, branchID).First(&module, moduleID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, utils.ErrorResponse{
				Status:  http.StatusNotFound,
				Message: "Module not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to retrieve module",
		})
		return
	}
	var permissions []models.Permission
	if err := masterDB.Where("business_id = ? AND branch_id = ? AND module_id = ?", businessID, branchID, moduleID).Find(&permissions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to retrieve permissions",
		})
		return
	}

	permList := make([]dtos.PermissionRes, 0, len(permissions))
	for _, perm := range permissions {
		permList = append(permList, dtos.PermissionRes{
			ID:         perm.ID,
			ModuleID:   perm.ModuleID,
			ModuleName: module.Name,
			Action:     perm.Action,
			BusinessId: perm.BusinessId,
			BranchId:   perm.BranchId,
			CreatedAt:  perm.CreatedAt,
			UpdatedAt:  perm.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Permissions retrieved successfully",
		Data: map[string]interface{}{
			"module_id":   module.ID,
			"module_name": module.Name,
			"permissions": permList,
			"total":       len(permList),
		},
	})
}
