package models

import "time"

type Subscription struct {
	ID         uint       `json:"id" gorm:"primaryKey"`
	BusinessId uint       `json:"business_id" binding:"required"`
	ApiKey     string     `json:"api_key"`
	SecretKey  string     `json:"secret_key"`
	BranchId   uint       `json:"branch_id"`
	ExpiryDate time.Time  `json:"expiry_date"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}
