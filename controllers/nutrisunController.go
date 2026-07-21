package controllers

import (
	"fmt"
	"net/http"
	"strings"

	"Shamas/nutrisun/config"
	"Shamas/nutrisun/dtos"
	"Shamas/nutrisun/middleware"
	"Shamas/nutrisun/models"
	"Shamas/nutrisun/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var validMealPreferences = map[string]bool{
	"veg":     true,
	"egg":     true,
	"non_veg": true,
}

var validAddressTypes = map[string]bool{
	"default":   true,
	"breakfast": true,
	"lunch":     true,
	"dinner":    true,
}

var nutriSunModuleDefinitions = []struct {
	Name        string
	DisplayName string
	Description string
}{
	{Name: "customer", DisplayName: "Customer Management", Description: "Manage NutriSun customers"},
	{Name: "customer_address", DisplayName: "Customer Address Management", Description: "Manage NutriSun customer addresses"},
}

func CreateCustomer(c *gin.Context, masterDB *gorm.DB) {
	var req dtos.CustomerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Status: http.StatusBadRequest, Message: "Invalid request: " + err.Error()})
		return
	}

	authContext, tenantDB, ok := getNutriSunContext(c)
	if !ok {
		return
	}

	preference, status, ok := validateCustomerCoreFields(c, req.DefaultMealPreference, req.Status)
	if !ok {
		return
	}

	if err := ensureCustomerUnique(tenantDB, authContext.BusinessID, authContext.BranchID, 0, req.PrimaryPhone, req.Email); err != nil {
		c.JSON(http.StatusConflict, utils.ErrorResponse{Status: http.StatusConflict, Message: err.Error()})
		return
	}

	customerCode, err := generateNextCustomerCode(tenantDB, authContext.BranchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Failed to generate customer code"})
		return
	}

	customer := models.Customer{
		CustomerCode:          customerCode,
		FullName:              strings.TrimSpace(req.FullName),
		PrimaryPhone:          strings.TrimSpace(req.PrimaryPhone),
		AlternatePhone:        strings.TrimSpace(req.AlternatePhone),
		Email:                 strings.TrimSpace(req.Email),
		Gender:                defaultString(normalize(req.Gender), "unknown"),
		DefaultMealPreference: preference,
		Status:                status,
		Notes:                 strings.TrimSpace(req.Notes),
		BusinessId:            authContext.BusinessID,
		BranchId:              authContext.BranchID,
		CreatedBy:             authContext.UserID,
		UpdatedBy:             authContext.UserID,
	}

	if strings.TrimSpace(req.Password) != "" {
		hashedPassword, err := utils.HashPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Failed to hash password"})
			return
		}
		customer.Password = hashedPassword
	}

	if err := tenantDB.Create(&customer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Failed to create customer: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, utils.SuccessResponse{
		Status:  http.StatusCreated,
		Message: "Customer created successfully",
		Data: map[string]interface{}{
			"customer": buildCustomerRes(customer),
		},
	})
}

func RegisterCustomer(c *gin.Context, masterDB *gorm.DB) {
	var req dtos.CustomerRegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Status: http.StatusBadRequest, Message: "Invalid request: " + err.Error()})
		return
	}

	preference := normalize(req.DefaultMealPreference)
	if !validMealPreferences[preference] {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Status: http.StatusBadRequest, Message: "default_meal_preference must be veg, egg, or non_veg"})
		return
	}

	var business models.Business
	if err := masterDB.First(&business, req.BusinessID).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{Status: http.StatusNotFound, Message: "Business not found"})
		return
	}

	var branch models.Branch
	if err := masterDB.Where("id = ? AND business_id = ?", req.BranchID, req.BusinessID).First(&branch).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{Status: http.StatusNotFound, Message: "Branch not found"})
		return
	}

	tenantDB, dbName := resolveTenantDBFromBusiness(c, business)
	if tenantDB == nil {
		return
	}

	if err := ensureCustomerUnique(tenantDB, req.BusinessID, req.BranchID, 0, req.PrimaryPhone, req.Email); err != nil {
		c.JSON(http.StatusConflict, utils.ErrorResponse{Status: http.StatusConflict, Message: err.Error()})
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Failed to hash password"})
		return
	}

	customerCode, err := generateNextCustomerCode(tenantDB, req.BranchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Failed to generate customer code"})
		return
	}

	customer := models.Customer{
		CustomerCode:          customerCode,
		FullName:              strings.TrimSpace(req.FullName),
		PrimaryPhone:          strings.TrimSpace(req.PrimaryPhone),
		AlternatePhone:        strings.TrimSpace(req.AlternatePhone),
		Email:                 strings.TrimSpace(req.Email),
		Password:              hashedPassword,
		Gender:                defaultString(normalize(req.Gender), "unknown"),
		DefaultMealPreference: preference,
		Status:                "active",
		Notes:                 strings.TrimSpace(req.Notes),
		BusinessId:            req.BusinessID,
		BranchId:              req.BranchID,
	}

	if err := tenantDB.Create(&customer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Failed to register customer: " + err.Error()})
		return
	}

	token, err := utils.GenerateScopedJWT(customer.ID, 0, customer.BusinessId, customer.BranchId, customer.PrimaryPhone, dbName, "customer")
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Failed to generate customer token"})
		return
	}

	c.JSON(http.StatusCreated, utils.SuccessResponse{
		Status:  http.StatusCreated,
		Message: "Customer registered successfully",
		Data: dtos.CustomerAuthRes{
			Customer: buildCustomerRes(customer),
			Token:    token,
		},
	})
}

func LoginCustomer(c *gin.Context, masterDB *gorm.DB) {
	var req dtos.CustomerLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Status: http.StatusBadRequest, Message: "Invalid request: " + err.Error()})
		return
	}

	var business models.Business
	if err := masterDB.First(&business, req.BusinessID).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{Status: http.StatusNotFound, Message: "Business not found"})
		return
	}

	tenantDB, dbName := resolveTenantDBFromBusiness(c, business)
	if tenantDB == nil {
		return
	}

	var customer models.Customer
	if err := tenantDB.Where("business_id = ? AND LOWER(primary_phone) = LOWER(?) AND status = ?", req.BusinessID, strings.TrimSpace(req.Phone), "active").First(&customer).Error; err != nil {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{Status: http.StatusUnauthorized, Message: "Invalid phone number or password"})
		return
	}

	if strings.TrimSpace(customer.Password) == "" || !utils.CheckPassword(req.Password, customer.Password) {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{Status: http.StatusUnauthorized, Message: "Invalid phone number or password"})
		return
	}

	token, err := utils.GenerateScopedJWT(customer.ID, 0, customer.BusinessId, customer.BranchId, customer.PrimaryPhone, dbName, "customer")
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Failed to generate customer token"})
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Customer login successful",
		Data: dtos.CustomerAuthRes{
			Customer: buildCustomerRes(customer),
			Token:    token,
		},
	})
}

func GetCustomerProfile(c *gin.Context, masterDB *gorm.DB) {
	customer, tenantDB, ok := getCustomerContext(c)
	if !ok {
		return
	}

	addresses, err := listCustomerAddresses(tenantDB, customer.ID, customer.BusinessId, customer.BranchId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Failed to retrieve customer addresses"})
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Customer profile retrieved successfully",
		Data: map[string]interface{}{
			"customer":  buildCustomerRes(*customer),
			"addresses": addresses,
		},
	})
}

func UpdateCustomerProfile(c *gin.Context, masterDB *gorm.DB) {
	var req dtos.CustomerProfileUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Status: http.StatusBadRequest, Message: "Invalid request: " + err.Error()})
		return
	}

	customer, tenantDB, ok := getCustomerContext(c)
	if !ok {
		return
	}

	if err := applyCustomerUpdates(tenantDB, customer, req.FullName, req.PrimaryPhone, req.AlternatePhone, req.Email, req.Gender, req.DefaultMealPreference, req.Notes, nil, customer.ID); err != nil {
		if err == gorm.ErrDuplicatedKey || strings.Contains(strings.ToLower(err.Error()), "already exists") {
			c.JSON(http.StatusConflict, utils.ErrorResponse{Status: http.StatusConflict, Message: err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Status: http.StatusBadRequest, Message: err.Error()})
		return
	}

	if err := tenantDB.Save(customer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Failed to update customer profile: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Customer profile updated successfully",
		Data: map[string]interface{}{
			"customer": buildCustomerRes(*customer),
		},
	})
}

func GetMyCustomerAddresses(c *gin.Context, masterDB *gorm.DB) {
	customer, tenantDB, ok := getCustomerContext(c)
	if !ok {
		return
	}

	addresses, err := listCustomerAddresses(tenantDB, customer.ID, customer.BusinessId, customer.BranchId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Failed to retrieve addresses"})
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{Status: http.StatusOK, Message: "Addresses retrieved successfully", Data: map[string]interface{}{"addresses": addresses}})
}

func CreateMyCustomerAddress(c *gin.Context, masterDB *gorm.DB) {
	customer, _, ok := getCustomerContext(c)
	if !ok {
		return
	}
	createAddressForCustomer(c, customer.ID, customer.BusinessId, customer.BranchId, customer.ID)
}

func UpdateMyCustomerAddress(c *gin.Context, masterDB *gorm.DB) {
	customer, _, ok := getCustomerContext(c)
	if !ok {
		return
	}
	updateAddressForCustomer(c, customer.ID, customer.BusinessId, customer.BranchId, customer.ID)
}

func DeleteMyCustomerAddress(c *gin.Context, masterDB *gorm.DB) {
	customer, _, ok := getCustomerContext(c)
	if !ok {
		return
	}
	deleteAddressForCustomer(c, customer.ID, customer.BusinessId, customer.BranchId)
}

func GetAllCustomers(c *gin.Context, masterDB *gorm.DB) {
	authContext, tenantDB, ok := getNutriSunContext(c)
	if !ok {
		return
	}

	page := utils.StringToInt(c.DefaultQuery("page", "1"))
	perPage := utils.StringToInt(c.DefaultQuery("per_page", "10"))
	search := strings.TrimSpace(c.Query("search"))
	status := normalize(c.Query("status"))

	query := tenantDB.Model(&models.Customer{}).Where("business_id = ? AND branch_id = ?", authContext.BusinessID, authContext.BranchID)
	if search != "" {
		like := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(full_name) LIKE ? OR LOWER(customer_code) LIKE ? OR LOWER(primary_phone) LIKE ? OR LOWER(email) LIKE ?", like, like, like, like)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Failed to count customers"})
		return
	}

	var customers []models.Customer
	offset := (page - 1) * perPage
	if err := query.Order("created_at DESC").Offset(offset).Limit(perPage).Find(&customers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Failed to retrieve customers"})
		return
	}

	results := make([]dtos.CustomerRes, 0, len(customers))
	for _, customer := range customers {
		results = append(results, buildCustomerRes(customer))
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Customers retrieved successfully",
		Data: map[string]interface{}{
			"customers": results,
			"pagination": map[string]interface{}{
				"total_records": total,
				"page":          page,
				"per_page":      perPage,
			},
		},
	})
}

func GetCustomerByID(c *gin.Context, masterDB *gorm.DB) {
	authContext, tenantDB, ok := getNutriSunContext(c)
	if !ok {
		return
	}

	customer, err := findCustomerByID(tenantDB, authContext.BusinessID, authContext.BranchID, utils.StringToUint(c.Param("id")))
	if err != nil {
		handleTenantRecordError(c, "Customer not found", err)
		return
	}

	addresses, err := listCustomerAddresses(tenantDB, customer.ID, authContext.BusinessID, authContext.BranchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Failed to retrieve addresses"})
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Customer retrieved successfully",
		Data: map[string]interface{}{
			"customer":  buildCustomerRes(*customer),
			"addresses": addresses,
		},
	})
}

func UpdateCustomer(c *gin.Context, masterDB *gorm.DB) {
	var req dtos.CustomerAdminUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Status: http.StatusBadRequest, Message: "Invalid request: " + err.Error()})
		return
	}

	authContext, tenantDB, ok := getNutriSunContext(c)
	if !ok {
		return
	}

	customer, err := findCustomerByID(tenantDB, authContext.BusinessID, authContext.BranchID, utils.StringToUint(c.Param("id")))
	if err != nil {
		handleTenantRecordError(c, "Customer not found", err)
		return
	}

	if err := applyCustomerUpdates(tenantDB, customer, req.FullName, req.PrimaryPhone, req.AlternatePhone, req.Email, req.Gender, req.DefaultMealPreference, req.Notes, stringPtrIfNotEmpty(req.Status), authContext.UserID); err != nil {
		if err == gorm.ErrDuplicatedKey || strings.Contains(strings.ToLower(err.Error()), "already exists") {
			c.JSON(http.StatusConflict, utils.ErrorResponse{Status: http.StatusConflict, Message: err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Status: http.StatusBadRequest, Message: err.Error()})
		return
	}

	if err := tenantDB.Save(customer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Failed to update customer: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Customer updated successfully",
		Data: map[string]interface{}{
			"customer": buildCustomerRes(*customer),
		},
	})
}

func UpdateCustomerStatus(c *gin.Context, masterDB *gorm.DB) {
	var req dtos.CustomerStatusUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Status: http.StatusBadRequest, Message: "Invalid request: " + err.Error()})
		return
	}

	authContext, tenantDB, ok := getNutriSunContext(c)
	if !ok {
		return
	}

	customer, err := findCustomerByID(tenantDB, authContext.BusinessID, authContext.BranchID, utils.StringToUint(c.Param("id")))
	if err != nil {
		handleTenantRecordError(c, "Customer not found", err)
		return
	}

	status := normalize(req.Status)
	if status != "active" && status != "inactive" {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Status: http.StatusBadRequest, Message: "status must be active or inactive"})
		return
	}

	customer.Status = status
	customer.UpdatedBy = authContext.UserID
	if err := tenantDB.Save(customer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Failed to update customer status: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Customer status updated successfully",
		Data: map[string]interface{}{
			"customer": buildCustomerRes(*customer),
		},
	})
}

func SetCustomerPassword(c *gin.Context, masterDB *gorm.DB) {
	var req dtos.CustomerSetPasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Status: http.StatusBadRequest, Message: "Invalid request: " + err.Error()})
		return
	}

	authContext, tenantDB, ok := getNutriSunContext(c)
	if !ok {
		return
	}

	customer, err := findCustomerByID(tenantDB, authContext.BusinessID, authContext.BranchID, utils.StringToUint(c.Param("id")))
	if err != nil {
		handleTenantRecordError(c, "Customer not found", err)
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Failed to hash password"})
		return
	}

	customer.Password = hashedPassword
	customer.UpdatedBy = authContext.UserID
	if err := tenantDB.Save(customer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Failed to set customer password: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{Status: http.StatusOK, Message: "Customer password updated successfully"})
}

func DeleteCustomer(c *gin.Context, masterDB *gorm.DB) {
	authContext, tenantDB, ok := getNutriSunContext(c)
	if !ok {
		return
	}

	customer, err := findCustomerByID(tenantDB, authContext.BusinessID, authContext.BranchID, utils.StringToUint(c.Param("id")))
	if err != nil {
		handleTenantRecordError(c, "Customer not found", err)
		return
	}

	if err := tenantDB.Delete(customer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Failed to delete customer: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{Status: http.StatusOK, Message: "Customer deleted successfully"})
}

func CreateCustomerAddress(c *gin.Context, masterDB *gorm.DB) {
	authContext, _, ok := getNutriSunContext(c)
	if !ok {
		return
	}
	createAddressForCustomer(c, utils.StringToUint(c.Param("id")), authContext.BusinessID, authContext.BranchID, authContext.UserID)
}

func GetCustomerAddresses(c *gin.Context, masterDB *gorm.DB) {
	authContext, tenantDB, ok := getNutriSunContext(c)
	if !ok {
		return
	}

	customerID := utils.StringToUint(c.Param("id"))
	if _, err := findCustomerByID(tenantDB, authContext.BusinessID, authContext.BranchID, customerID); err != nil {
		handleTenantRecordError(c, "Customer not found", err)
		return
	}

	addresses, err := listCustomerAddresses(tenantDB, customerID, authContext.BusinessID, authContext.BranchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Failed to retrieve addresses"})
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{Status: http.StatusOK, Message: "Addresses retrieved successfully", Data: map[string]interface{}{"addresses": addresses}})
}

func UpdateCustomerAddress(c *gin.Context, masterDB *gorm.DB) {
	authContext, _, ok := getNutriSunContext(c)
	if !ok {
		return
	}
	updateAddressForCustomer(c, utils.StringToUint(c.Param("id")), authContext.BusinessID, authContext.BranchID, authContext.UserID)
}

func DeleteCustomerAddress(c *gin.Context, masterDB *gorm.DB) {
	authContext, _, ok := getNutriSunContext(c)
	if !ok {
		return
	}
	deleteAddressForCustomer(c, utils.StringToUint(c.Param("id")), authContext.BusinessID, authContext.BranchID)
}

func BootstrapNutriSunModules(c *gin.Context, masterDB *gorm.DB) {
	authContext, _, ok := getNutriSunContext(c)
	if !ok {
		return
	}

	var req struct {
		RoleID uint `json:"role_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Status: http.StatusBadRequest, Message: "Invalid request: " + err.Error()})
		return
	}

	var role models.Role
	if err := masterDB.Where("id = ? AND business_id = ? AND branch_id = ?", req.RoleID, authContext.BusinessID, authContext.BranchID).First(&role).Error; err != nil {
		handleTenantRecordError(c, "Role not found", err)
		return
	}

	createdModules := make([]dtos.ModuleRes, 0, len(nutriSunModuleDefinitions))
	err := masterDB.Transaction(func(tx *gorm.DB) error {
		for _, definition := range nutriSunModuleDefinitions {
			module := models.Module{
				Name:        definition.Name,
				DisplayName: definition.DisplayName,
				Description: definition.Description,
				BusinessId:  authContext.BusinessID,
				BranchId:    authContext.BranchID,
			}
			if err := tx.Where("business_id = ? AND branch_id = ? AND name = ?", authContext.BusinessID, authContext.BranchID, definition.Name).FirstOrCreate(&module, module).Error; err != nil {
				return err
			}
			createdModules = append(createdModules, dtos.ModuleRes{
				ID:          module.ID,
				Name:        module.Name,
				DisplayName: module.DisplayName,
				Description: module.Description,
				BusinessId:  module.BusinessId,
				BranchId:    module.BranchId,
				CreatedAt:   module.CreatedAt,
				UpdatedAt:   module.UpdatedAt,
			})
			for _, action := range []string{"get", "create", "update", "delete"} {
				permission := models.Permission{
					RoleID:     role.ID,
					ModuleID:   module.ID,
					Action:     action,
					BusinessId: authContext.BusinessID,
					BranchId:   authContext.BranchID,
				}
				if err := tx.Where("business_id = ? AND branch_id = ? AND role_id = ? AND module_id = ? AND action = ?", authContext.BusinessID, authContext.BranchID, role.ID, module.ID, action).FirstOrCreate(&permission, permission).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Failed to bootstrap customer modules: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Customer modules bootstrapped successfully",
		Data: map[string]interface{}{
			"role_id":   role.ID,
			"role_name": role.Name,
			"modules":   createdModules,
		},
	})
}

func getNutriSunContext(c *gin.Context) (*utils.UserContext, *gorm.DB, bool) {
	authContext, ok := utils.GetCurrentUserContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{Status: http.StatusUnauthorized, Message: "User context not found"})
		return nil, nil, false
	}

	tenantDB := middleware.GetTenantSQLDB(c)
	if tenantDB == nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Tenant database connection not found"})
		return nil, nil, false
	}

	return authContext, tenantDB, true
}

func getCustomerContext(c *gin.Context) (*models.Customer, *gorm.DB, bool) {
	customer := middleware.GetCustomer(c)
	if customer == nil {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{Status: http.StatusUnauthorized, Message: "Customer not found in token"})
		return nil, nil, false
	}

	tenantDB := middleware.GetTenantSQLDB(c)
	if tenantDB == nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Tenant database connection not found"})
		return nil, nil, false
	}

	return customer, tenantDB, true
}

func resolveTenantDBFromBusiness(c *gin.Context, business models.Business) (*gorm.DB, string) {
	dbName := business.DB
	if strings.TrimSpace(dbName) == "" {
		dbName = utils.GetDefaultTenantDBName()
	}
	tenantDB := config.ConnectTenantSQLDB(dbName)
	if tenantDB == nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Failed to connect tenant database"})
		return nil, ""
	}
	return tenantDB, dbName
}

func findCustomerByID(db *gorm.DB, businessID, branchID, customerID uint) (*models.Customer, error) {
	var customer models.Customer
	if err := db.Where("id = ? AND business_id = ? AND branch_id = ?", customerID, businessID, branchID).First(&customer).Error; err != nil {
		return nil, err
	}
	return &customer, nil
}

func generateNextCustomerCode(db *gorm.DB, branchID uint) (string, error) {
	var latest models.Customer
	if err := db.Order("id DESC").First(&latest).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Sprintf("NSC-%04d-0001", branchID), nil
		}
		return "", err
	}

	var sequence int
	_, _ = fmt.Sscanf(latest.CustomerCode, "NSC-%*d-%d", &sequence)
	if sequence <= 0 {
		sequence = int(latest.ID)
	}
	return fmt.Sprintf("NSC-%04d-%04d", branchID, sequence+1), nil
}

func listCustomerAddresses(tenantDB *gorm.DB, customerID, businessID, branchID uint) ([]dtos.CustomerAddressRes, error) {
	var addresses []models.CustomerAddress
	if err := tenantDB.Where("customer_id = ? AND business_id = ? AND branch_id = ?", customerID, businessID, branchID).
		Order("is_primary DESC, created_at DESC").
		Find(&addresses).Error; err != nil {
		return nil, err
	}

	results := make([]dtos.CustomerAddressRes, 0, len(addresses))
	for _, address := range addresses {
		results = append(results, buildCustomerAddressRes(address))
	}
	return results, nil
}

func createAddressForCustomer(c *gin.Context, customerID, businessID, branchID, actorID uint) {
	var req dtos.CustomerAddressReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Status: http.StatusBadRequest, Message: "Invalid request: " + err.Error()})
		return
	}

	tenantDB := middleware.GetTenantSQLDB(c)
	if tenantDB == nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Tenant database connection not found"})
		return
	}

	if _, err := findCustomerByID(tenantDB, businessID, branchID, customerID); err != nil {
		handleTenantRecordError(c, "Customer not found", err)
		return
	}

	addressType := normalize(req.AddressType)
	if !validAddressTypes[addressType] {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Status: http.StatusBadRequest, Message: "address_type must be default, breakfast, lunch, or dinner"})
		return
	}

	err := tenantDB.Transaction(func(tx *gorm.DB) error {
		if req.IsPrimary {
			if err := tx.Model(&models.CustomerAddress{}).
				Where("customer_id = ? AND business_id = ? AND branch_id = ?", customerID, businessID, branchID).
				Update("is_primary", false).Error; err != nil {
				return err
			}
		}

		address := models.CustomerAddress{
			CustomerID:     customerID,
			AddressType:    addressType,
			AddressLine:    strings.TrimSpace(req.AddressLine),
			Area:           strings.TrimSpace(req.Area),
			Landmark:       strings.TrimSpace(req.Landmark),
			ContactPerson:  strings.TrimSpace(req.ContactPerson),
			AlternatePhone: strings.TrimSpace(req.AlternatePhone),
			Latitude:       strings.TrimSpace(req.Latitude),
			Longitude:      strings.TrimSpace(req.Longitude),
			DeliveryNotes:  strings.TrimSpace(req.DeliveryNotes),
			IsPrimary:      req.IsPrimary,
			BusinessId:     businessID,
			BranchId:       branchID,
			CreatedBy:      actorID,
			UpdatedBy:      actorID,
		}
		if err := tx.Create(&address).Error; err != nil {
			return err
		}

		c.JSON(http.StatusCreated, utils.SuccessResponse{
			Status:  http.StatusCreated,
			Message: "Customer address created successfully",
			Data: map[string]interface{}{
				"address": buildCustomerAddressRes(address),
			},
		})
		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Failed to create customer address: " + err.Error()})
	}
}

func updateAddressForCustomer(c *gin.Context, customerID, businessID, branchID, actorID uint) {
	var req dtos.CustomerAddressUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Status: http.StatusBadRequest, Message: "Invalid request: " + err.Error()})
		return
	}

	tenantDB := middleware.GetTenantSQLDB(c)
	if tenantDB == nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Tenant database connection not found"})
		return
	}

	addressID := utils.StringToUint(c.Param("address_id"))
	var address models.CustomerAddress
	if err := tenantDB.Where("id = ? AND customer_id = ? AND business_id = ? AND branch_id = ?", addressID, customerID, businessID, branchID).First(&address).Error; err != nil {
		handleTenantRecordError(c, "Customer address not found", err)
		return
	}

	if strings.TrimSpace(req.AddressType) != "" {
		addressType := normalize(req.AddressType)
		if !validAddressTypes[addressType] {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{Status: http.StatusBadRequest, Message: "address_type must be default, breakfast, lunch, or dinner"})
			return
		}
		address.AddressType = addressType
	}
	if strings.TrimSpace(req.AddressLine) != "" {
		address.AddressLine = strings.TrimSpace(req.AddressLine)
	}
	if strings.TrimSpace(req.Area) != "" {
		address.Area = strings.TrimSpace(req.Area)
	}
	if strings.TrimSpace(req.Landmark) != "" {
		address.Landmark = strings.TrimSpace(req.Landmark)
	}
	if strings.TrimSpace(req.ContactPerson) != "" {
		address.ContactPerson = strings.TrimSpace(req.ContactPerson)
	}
	if strings.TrimSpace(req.AlternatePhone) != "" {
		address.AlternatePhone = strings.TrimSpace(req.AlternatePhone)
	}
	if strings.TrimSpace(req.Latitude) != "" {
		address.Latitude = strings.TrimSpace(req.Latitude)
	}
	if strings.TrimSpace(req.Longitude) != "" {
		address.Longitude = strings.TrimSpace(req.Longitude)
	}
	if strings.TrimSpace(req.DeliveryNotes) != "" {
		address.DeliveryNotes = strings.TrimSpace(req.DeliveryNotes)
	}

	err := tenantDB.Transaction(func(tx *gorm.DB) error {
		if req.IsPrimary != nil && *req.IsPrimary {
			if err := tx.Model(&models.CustomerAddress{}).
				Where("customer_id = ? AND business_id = ? AND branch_id = ?", customerID, businessID, branchID).
				Update("is_primary", false).Error; err != nil {
				return err
			}
			address.IsPrimary = true
		} else if req.IsPrimary != nil {
			address.IsPrimary = false
		}

		address.UpdatedBy = actorID
		return tx.Save(&address).Error
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Failed to update customer address: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Customer address updated successfully",
		Data: map[string]interface{}{
			"address": buildCustomerAddressRes(address),
		},
	})
}

func deleteAddressForCustomer(c *gin.Context, customerID, businessID, branchID uint) {
	tenantDB := middleware.GetTenantSQLDB(c)
	if tenantDB == nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Tenant database connection not found"})
		return
	}

	addressID := utils.StringToUint(c.Param("address_id"))
	var address models.CustomerAddress
	if err := tenantDB.Where("id = ? AND customer_id = ? AND business_id = ? AND branch_id = ?", addressID, customerID, businessID, branchID).First(&address).Error; err != nil {
		handleTenantRecordError(c, "Customer address not found", err)
		return
	}

	if err := tenantDB.Delete(&address).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Failed to delete customer address: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{Status: http.StatusOK, Message: "Customer address deleted successfully"})
}

func validateCustomerCoreFields(c *gin.Context, mealPreference, status string) (string, string, bool) {
	preference := normalize(mealPreference)
	if !validMealPreferences[preference] {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Status: http.StatusBadRequest, Message: "default_meal_preference must be veg, egg, or non_veg"})
		return "", "", false
	}

	normalizedStatus := normalize(status)
	if normalizedStatus == "" {
		normalizedStatus = "active"
	}
	if normalizedStatus != "active" && normalizedStatus != "inactive" {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Status: http.StatusBadRequest, Message: "status must be active or inactive"})
		return "", "", false
	}
	return preference, normalizedStatus, true
}

func ensureCustomerUnique(tenantDB *gorm.DB, businessID, branchID, customerID uint, primaryPhone, email string) error {
	phone := strings.TrimSpace(primaryPhone)
	if phone != "" {
		var count int64
		query := tenantDB.Model(&models.Customer{}).Where("business_id = ? AND branch_id = ? AND LOWER(primary_phone) = LOWER(?)", businessID, branchID, phone)
		if customerID > 0 {
			query = query.Where("id <> ?", customerID)
		}
		if err := query.Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return fmt.Errorf("customer phone number already exists")
		}
	}

	email = strings.TrimSpace(email)
	if email != "" {
		var count int64
		query := tenantDB.Model(&models.Customer{}).Where("business_id = ? AND branch_id = ? AND LOWER(email) = LOWER(?)", businessID, branchID, email)
		if customerID > 0 {
			query = query.Where("id <> ?", customerID)
		}
		if err := query.Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return fmt.Errorf("customer email already exists")
		}
	}

	return nil
}

func applyCustomerUpdates(tenantDB *gorm.DB, customer *models.Customer, fullName, primaryPhone, alternatePhone, email, gender, mealPreference, notes string, status *string, actorID uint) error {
	if strings.TrimSpace(primaryPhone) != "" || strings.TrimSpace(email) != "" {
		if err := ensureCustomerUnique(tenantDB, customer.BusinessId, customer.BranchId, customer.ID, defaultString(strings.TrimSpace(primaryPhone), customer.PrimaryPhone), defaultString(strings.TrimSpace(email), customer.Email)); err != nil {
			return err
		}
	}

	if strings.TrimSpace(fullName) != "" {
		customer.FullName = strings.TrimSpace(fullName)
	}
	if strings.TrimSpace(primaryPhone) != "" {
		customer.PrimaryPhone = strings.TrimSpace(primaryPhone)
	}
	if strings.TrimSpace(alternatePhone) != "" {
		customer.AlternatePhone = strings.TrimSpace(alternatePhone)
	}
	if strings.TrimSpace(email) != "" {
		customer.Email = strings.TrimSpace(email)
	}
	if strings.TrimSpace(gender) != "" {
		customer.Gender = normalize(gender)
	}
	if strings.TrimSpace(mealPreference) != "" {
		preference := normalize(mealPreference)
		if !validMealPreferences[preference] {
			return fmt.Errorf("default_meal_preference must be veg, egg, or non_veg")
		}
		customer.DefaultMealPreference = preference
	}
	if strings.TrimSpace(notes) != "" {
		customer.Notes = strings.TrimSpace(notes)
	}
	if status != nil {
		normalizedStatus := normalize(*status)
		if normalizedStatus != "active" && normalizedStatus != "inactive" {
			return fmt.Errorf("status must be active or inactive")
		}
		customer.Status = normalizedStatus
	}
	customer.UpdatedBy = actorID
	return nil
}

func normalize(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func stringPtrIfNotEmpty(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	trimmed := value
	return &trimmed
}

func handleTenantRecordError(c *gin.Context, notFoundMessage string, err error) {
	if err == gorm.ErrRecordNotFound {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{Status: http.StatusNotFound, Message: notFoundMessage})
		return
	}
	c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: err.Error()})
}

func buildCustomerRes(customer models.Customer) dtos.CustomerRes {
	return dtos.CustomerRes{
		ID:                    customer.ID,
		CustomerCode:          customer.CustomerCode,
		FullName:              customer.FullName,
		PrimaryPhone:          customer.PrimaryPhone,
		AlternatePhone:        customer.AlternatePhone,
		Email:                 customer.Email,
		Gender:                customer.Gender,
		DefaultMealPreference: customer.DefaultMealPreference,
		Status:                customer.Status,
		Notes:                 customer.Notes,
		BusinessId:            customer.BusinessId,
		BranchId:              customer.BranchId,
		CreatedAt:             customer.CreatedAt,
		UpdatedAt:             customer.UpdatedAt,
	}
}

func buildCustomerAddressRes(address models.CustomerAddress) dtos.CustomerAddressRes {
	return dtos.CustomerAddressRes{
		ID:             address.ID,
		CustomerID:     address.CustomerID,
		AddressType:    address.AddressType,
		AddressLine:    address.AddressLine,
		Area:           address.Area,
		Landmark:       address.Landmark,
		ContactPerson:  address.ContactPerson,
		AlternatePhone: address.AlternatePhone,
		Latitude:       address.Latitude,
		Longitude:      address.Longitude,
		DeliveryNotes:  address.DeliveryNotes,
		IsPrimary:      address.IsPrimary,
		CreatedAt:      address.CreatedAt,
		UpdatedAt:      address.UpdatedAt,
	}
}
