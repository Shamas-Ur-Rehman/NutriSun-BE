package models

import "time"

type MenuMonth struct {
	ID         uint       `json:"id" gorm:"primaryKey"`
	Month      int        `json:"month" gorm:"not null"`
	Year       int        `json:"year" gorm:"not null"`
	Name       string     `json:"name" gorm:"size:120;not null"`
	Status     string     `json:"status" gorm:"type:varchar(20);not null;default:'draft';check:status IN ('draft','published','archived')"`
	Version    int        `json:"version" gorm:"not null;default:1"`
	Notes      string     `json:"notes"`
	BusinessId uint       `json:"business_id" gorm:"index;not null"`
	BranchId   uint       `json:"branch_id" gorm:"index;not null"`
	CreatedBy  uint       `json:"created_by"`
	UpdatedBy  uint       `json:"updated_by"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

type MenuDay struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	MenuMonthID uint       `json:"menu_month_id" gorm:"index;not null"`
	MenuDate    time.Time  `json:"menu_date" gorm:"type:date;not null;index"`
	DayName     string     `json:"day_name" gorm:"size:20;not null"`
	BusinessId  uint       `json:"business_id" gorm:"index;not null"`
	BranchId    uint       `json:"branch_id" gorm:"index;not null"`
	CreatedBy   uint       `json:"created_by"`
	UpdatedBy   uint       `json:"updated_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

type MenuDayItem struct {
	ID             uint       `json:"id" gorm:"primaryKey"`
	MenuDayID      uint       `json:"menu_day_id" gorm:"index;not null"`
	MealSlot       string     `json:"meal_slot" gorm:"type:varchar(20);not null;check:meal_slot IN ('breakfast','lunch','dinner')"`
	ItemName       string     `json:"item_name" gorm:"size:150;not null"`
	Classification string     `json:"classification" gorm:"type:varchar(20);not null;check:classification IN ('veg','egg','non_veg')"`
	DisplayOrder   int        `json:"display_order" gorm:"not null;default:1"`
	BusinessId     uint       `json:"business_id" gorm:"index;not null"`
	BranchId       uint       `json:"branch_id" gorm:"index;not null"`
	CreatedBy      uint       `json:"created_by"`
	UpdatedBy      uint       `json:"updated_by"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}
