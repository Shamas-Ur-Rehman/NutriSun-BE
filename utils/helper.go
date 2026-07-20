package utils

import (
	"errors"
	"fmt"
	"math"
	"os"
	"time"

	"Shamas/nutrisun/dtos"
	"Shamas/nutrisun/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetUserByID(db *gorm.DB, userID uint) (*models.User, error) {
	var user models.User
	if err := db.Where("id = ? AND is_active = ?", userID, true).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func GetUserByEmail(db *gorm.DB, email string) (*models.User, error) {
	var user models.User
	if err := db.Where("email = ? AND is_active = ?", email, true).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}
func GetUserByEmployeeID(db *gorm.DB, employeeID string) (*models.User, error) {
	var user models.User
	if err := db.Where("employee_id = ? AND is_active = ?", employeeID, true).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}
func GetUserResponse(user *models.User) dtos.UserRes {
	return dtos.UserRes{
		Email:      user.Email,
		FullName:   user.FullName,
		DocumentId: user.DocumentId,
		Address:    user.Address,
		RoleID:     user.RoleID,
		BranchId:   user.BranchId,
		BusinessId: user.BusinessId,
		FCMToken:   user.FCMToken,
		IsActive:   user.IsActive,
	}
}

func GetUsersByBusiness(db *gorm.DB, businessID uint) ([]models.User, error) {
	var users []models.User
	if err := db.Where("business_id = ? AND is_active = ?", businessID, true).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func GetUsersByBranch(db *gorm.DB, branchID uint) ([]models.User, error) {
	var users []models.User
	if err := db.Where("branch_id = ? AND is_active = ?", branchID, true).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func GetUsersByRole(db *gorm.DB, roleID uint) ([]models.User, error) {
	var users []models.User
	if err := db.Where("role_id = ? AND is_active = ?", roleID, true).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func GetTenantDbName(c *gin.Context) (string, bool) {
	if dbName, exist := c.Get("tenant_db"); exist {
		return dbName.(string), true
	}
	dbName, exist := c.Get("db_name")
	if !exist {
		return "", false
	}
	return dbName.(string), true
}

type PaginationResult struct {
	Data         interface{} `json:"data"`
	TotalRecords int64       `json:"total_records"`
	TotalPages   int         `json:"total_pages"`
	Page         int         `json:"page"`
	PerPage      int         `json:"per_page"`
}

func PaginateMySQL(db *gorm.DB, model interface{}, filter map[string]interface{}, page int, perPage int, result interface{}) (*PaginationResult, error) {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 10
	}

	query := db.Model(model).Where(filter)

	var totalRecords int64
	if err := query.Count(&totalRecords).Error; err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(totalRecords) / float64(perPage)))
	offset := (page - 1) * perPage

	if err := query.Limit(perPage).Offset(offset).Find(result).Error; err != nil {
		return nil, err
	}

	return &PaginationResult{
		Data:         result,
		TotalRecords: totalRecords,
		TotalPages:   totalPages,
		Page:         page,
		PerPage:      perPage,
	}, nil
}

func StringToInt(s string) int {
	var result int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + int(c-'0')
		}
	}
	if result == 0 {
		return 1
	}
	return result
}

func GetDefaultTenantDBName() string {
	if name := firstNonEmptyEnv("TENANT_DB_NAME", "DB_NAME"); name != "" {
		return name
	}
	return "fms_master_db"
}

func firstNonEmptyEnv(keys ...string) string {
	for _, key := range keys {
		if value := os.Getenv(key); value != "" {
			return value
		}
	}
	return ""
}

func StringToUint(s string) uint {
	var result uint
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + uint(c-'0')
		}
	}
	return result
}
func GenerateNextEmployeeID(db *gorm.DB) (string, error) {
	var lastUser models.User
	err := db.Order("id DESC").First(&lastUser).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "IPEMP0001", nil
		}
		return "", err
	}
	if lastUser.EmployeeId == "" {
		return "IPEMP0001", nil
	}
	var lastNumber int
	_, err = fmt.Sscanf(lastUser.EmployeeId, "IPEMP%d", &lastNumber)
	if err != nil {
		return "IPEMP0001", nil
	}
	nextNumber := lastNumber + 1
	return fmt.Sprintf("IPEMP%04d", nextNumber), nil
}

func CalculateAge(dob time.Time) int {
	now := time.Now()
	age := now.Year() - dob.Year()
	if now.Month() < dob.Month() || (now.Month() == dob.Month() && now.Day() < dob.Day()) {
		age--
	}

	return age
}

func BuildUserResponse(user *models.User) dtos.UserRes {
	userRes := dtos.UserRes{
		ID:          user.ID,
		Email:       user.Email,
		FullName:    user.FullName,
		EmployeeId:  user.EmployeeId,
		Contact:     user.Contact,
		Gender:      user.Gender,
		DocumentId:  user.DocumentId,
		License:     user.License,
		Address:     user.Address,
		Nationality: user.Nationality,
		RoleID:      user.RoleID,
		RoleName:    user.Role.Name,
		BranchId:    user.BranchId,
		BusinessId:  user.BusinessId,
		FCMToken:    user.FCMToken,
		IsActive:    user.IsActive,
		IsStaff:     user.IsStaff,
		Service:     user.Service,
	}

	if user.DateOfBirth != nil {
		dobStr := user.DateOfBirth.Format("2006-01-02")
		userRes.DateOfBirth = &dobStr

		age := CalculateAge(*user.DateOfBirth)
		userRes.Age = &age
	}

	return userRes
}

type UserContext struct {
	UserID       uint
	BusinessID   uint
	BranchID     uint
	User         *models.User
	Business     *models.Business
	TenantDB     string
	IsAPIRequest bool
}

func GetCurrentUserContext(c *gin.Context) (*UserContext, bool) {
	businessID, businessExists := c.Get("business_id")
	branchID, branchExists := c.Get("branch_id")
	user, userExists := c.Get("user")

	if !businessExists || !branchExists || !userExists {
		return nil, false
	}

	userModel, ok := user.(*models.User)
	if !ok {
		return nil, false
	}

	context := &UserContext{
		UserID:     userModel.ID,
		BusinessID: businessID.(uint),
		BranchID:   branchID.(uint),
		User:       userModel,
	}

	if businessValue, exists := c.Get("business"); exists {
		if businessModel, ok := businessValue.(models.Business); ok {
			context.Business = &businessModel
		}
		if businessModel, ok := businessValue.(*models.Business); ok {
			context.Business = businessModel
		}
	}

	if tenantDBValue, exists := c.Get("tenant_db"); exists {
		if tenantDB, ok := tenantDBValue.(string); ok {
			context.TenantDB = tenantDB
		}
	}

	if isAPIValue, exists := c.Get("is_api_request"); exists {
		if isAPI, ok := isAPIValue.(bool); ok {
			context.IsAPIRequest = isAPI
		}
	}

	return context, true
}

func GetBusinessAndBranch(c *gin.Context) (*UserContext, bool) {
	return GetCurrentUserContext(c)
}
