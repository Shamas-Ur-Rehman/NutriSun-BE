package models

import "time"

type SubscriptionPlan struct {
	ID            uint       `json:"id" gorm:"primaryKey"`
	PlanName      string     `json:"plan_name" gorm:"size:150;not null"`
	DurationType  string     `json:"duration_type" gorm:"type:varchar(20);not null;check:duration_type IN ('weekly','monthly')"`
	MealCombo     string     `json:"meal_combo" gorm:"type:varchar(50);not null"`
	Price         float64    `json:"price" gorm:"type:numeric(12,2);not null"`
	EffectiveFrom time.Time  `json:"effective_from" gorm:"type:date;not null"`
	EffectiveTo   *time.Time `json:"effective_to,omitempty" gorm:"type:date"`
	IsActive      bool       `json:"is_active" gorm:"default:true"`
	Description   string     `json:"description"`
	BusinessId    uint       `json:"business_id" gorm:"index;not null"`
	BranchId      uint       `json:"branch_id" gorm:"index;not null"`
	CreatedBy     uint       `json:"created_by"`
	UpdatedBy     uint       `json:"updated_by"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}
