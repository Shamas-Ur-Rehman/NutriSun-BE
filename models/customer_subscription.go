package models

import "time"

type CustomerSubscription struct {
	ID                 uint       `json:"id" gorm:"primaryKey"`
	CustomerID         uint       `json:"customer_id" gorm:"index;not null"`
	SubscriptionPlanID uint       `json:"subscription_plan_id" gorm:"index;not null"`
	StartDate          time.Time  `json:"start_date" gorm:"type:date;not null"`
	EndDate            time.Time  `json:"end_date" gorm:"type:date;not null"`
	MealPreference     string     `json:"meal_preference" gorm:"type:varchar(20);not null;check:meal_preference IN ('veg','egg','non_veg')"`
	Status             string     `json:"status" gorm:"type:varchar(20);not null;default:'draft';check:status IN ('draft','active','paused','completed','cancelled')"`
	BillingStatus      string     `json:"billing_status" gorm:"type:varchar(20);not null;default:'pending';check:billing_status IN ('pending','invoiced','partially_paid','paid','cancelled')"`
	Notes              string     `json:"notes"`
	BusinessId         uint       `json:"business_id" gorm:"index;not null"`
	BranchId           uint       `json:"branch_id" gorm:"index;not null"`
	CreatedBy          uint       `json:"created_by"`
	UpdatedBy          uint       `json:"updated_by"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	DeletedAt          *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

type CustomerSubscriptionDay struct {
	ID                     uint       `json:"id" gorm:"primaryKey"`
	CustomerSubscriptionID uint       `json:"customer_subscription_id" gorm:"index;not null"`
	CustomerID             uint       `json:"customer_id" gorm:"index;not null"`
	ScheduleDate           time.Time  `json:"schedule_date" gorm:"type:date;not null;index"`
	DayName                string     `json:"day_name" gorm:"size:20;not null"`
	BreakfastRequired      bool       `json:"breakfast_required" gorm:"default:false"`
	LunchRequired          bool       `json:"lunch_required" gorm:"default:false"`
	DinnerRequired         bool       `json:"dinner_required" gorm:"default:false"`
	BreakfastSkipped       bool       `json:"breakfast_skipped" gorm:"default:false"`
	LunchSkipped           bool       `json:"lunch_skipped" gorm:"default:false"`
	DinnerSkipped          bool       `json:"dinner_skipped" gorm:"default:false"`
	MealPreference         string     `json:"meal_preference" gorm:"type:varchar(20);not null;check:meal_preference IN ('veg','egg','non_veg')"`
	MenuDayID              *uint      `json:"menu_day_id,omitempty" gorm:"index"`
	Status                 string     `json:"status" gorm:"type:varchar(20);not null;default:'scheduled';check:status IN ('scheduled','skipped','delivered','cancelled')"`
	BusinessId             uint       `json:"business_id" gorm:"index;not null"`
	BranchId               uint       `json:"branch_id" gorm:"index;not null"`
	CreatedBy              uint       `json:"created_by"`
	UpdatedBy              uint       `json:"updated_by"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
	DeletedAt              *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}
