package dtos

import "time"

type CustomerReq struct {
	FullName              string `json:"full_name" binding:"required"`
	PrimaryPhone          string `json:"primary_phone" binding:"required"`
	AlternatePhone        string `json:"alternate_phone"`
	Email                 string `json:"email" binding:"omitempty,email"`
	Gender                string `json:"gender"`
	DefaultMealPreference string `json:"default_meal_preference" binding:"required"`
	Status                string `json:"status"`
	Notes                 string `json:"notes"`
	Password              string `json:"password"`
}

type CustomerRegisterReq struct {
	BusinessID            uint   `json:"business_id" binding:"required"`
	BranchID              uint   `json:"branch_id" binding:"required"`
	FullName              string `json:"full_name" binding:"required"`
	PrimaryPhone          string `json:"primary_phone" binding:"required"`
	AlternatePhone        string `json:"alternate_phone"`
	Email                 string `json:"email" binding:"omitempty,email"`
	Password              string `json:"password" binding:"required,min=8"`
	Gender                string `json:"gender"`
	DefaultMealPreference string `json:"default_meal_preference" binding:"required"`
	Notes                 string `json:"notes"`
}

type CustomerLoginReq struct {
	BusinessID uint   `json:"business_id" binding:"required"`
	Phone      string `json:"phone" binding:"required"`
	Password   string `json:"password" binding:"required"`
}

type CustomerAdminUpdateReq struct {
	FullName              string `json:"full_name"`
	PrimaryPhone          string `json:"primary_phone"`
	AlternatePhone        string `json:"alternate_phone"`
	Email                 string `json:"email" binding:"omitempty,email"`
	Gender                string `json:"gender"`
	DefaultMealPreference string `json:"default_meal_preference"`
	Status                string `json:"status"`
	Notes                 string `json:"notes"`
}

type CustomerProfileUpdateReq struct {
	FullName              string `json:"full_name"`
	PrimaryPhone          string `json:"primary_phone"`
	AlternatePhone        string `json:"alternate_phone"`
	Email                 string `json:"email" binding:"omitempty,email"`
	Gender                string `json:"gender"`
	DefaultMealPreference string `json:"default_meal_preference"`
	Notes                 string `json:"notes"`
}

type CustomerStatusUpdateReq struct {
	Status string `json:"status" binding:"required"`
}

type CustomerSetPasswordReq struct {
	Password string `json:"password" binding:"required,min=8"`
}

type CustomerAuthRes struct {
	Customer CustomerRes `json:"customer"`
	Token    string      `json:"token"`
}

type CustomerRes struct {
	ID                    uint      `json:"id"`
	CustomerCode          string    `json:"customer_code"`
	FullName              string    `json:"full_name"`
	PrimaryPhone          string    `json:"primary_phone"`
	AlternatePhone        string    `json:"alternate_phone"`
	Email                 string    `json:"email"`
	Gender                string    `json:"gender"`
	DefaultMealPreference string    `json:"default_meal_preference"`
	Status                string    `json:"status"`
	Notes                 string    `json:"notes"`
	BusinessId            uint      `json:"business_id"`
	BranchId              uint      `json:"branch_id"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type CustomerAddressReq struct {
	AddressType    string `json:"address_type" binding:"required"`
	AddressLine    string `json:"address_line" binding:"required"`
	Area           string `json:"area"`
	Landmark       string `json:"landmark"`
	ContactPerson  string `json:"contact_person"`
	AlternatePhone string `json:"alternate_phone"`
	Latitude       string `json:"latitude"`
	Longitude      string `json:"longitude"`
	DeliveryNotes  string `json:"delivery_notes"`
	IsPrimary      bool   `json:"is_primary"`
}

type CustomerAddressUpdateReq struct {
	AddressType    string `json:"address_type"`
	AddressLine    string `json:"address_line"`
	Area           string `json:"area"`
	Landmark       string `json:"landmark"`
	ContactPerson  string `json:"contact_person"`
	AlternatePhone string `json:"alternate_phone"`
	Latitude       string `json:"latitude"`
	Longitude      string `json:"longitude"`
	DeliveryNotes  string `json:"delivery_notes"`
	IsPrimary      *bool  `json:"is_primary"`
}

type CustomerAddressRes struct {
	ID             uint      `json:"id"`
	CustomerID     uint      `json:"customer_id"`
	AddressType    string    `json:"address_type"`
	AddressLine    string    `json:"address_line"`
	Area           string    `json:"area"`
	Landmark       string    `json:"landmark"`
	ContactPerson  string    `json:"contact_person"`
	AlternatePhone string    `json:"alternate_phone"`
	Latitude       string    `json:"latitude"`
	Longitude      string    `json:"longitude"`
	DeliveryNotes  string    `json:"delivery_notes"`
	IsPrimary      bool      `json:"is_primary"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type SubscriptionPlanReq struct {
	PlanName      string  `json:"plan_name" binding:"required"`
	DurationType  string  `json:"duration_type" binding:"required"`
	MealCombo     string  `json:"meal_combo" binding:"required"`
	Price         float64 `json:"price" binding:"required"`
	EffectiveFrom string  `json:"effective_from" binding:"required"`
	EffectiveTo   string  `json:"effective_to"`
	IsActive      *bool   `json:"is_active"`
	Description   string  `json:"description"`
}

type SubscriptionPlanRes struct {
	ID            uint       `json:"id"`
	PlanName      string     `json:"plan_name"`
	DurationType  string     `json:"duration_type"`
	MealCombo     string     `json:"meal_combo"`
	Price         float64    `json:"price"`
	EffectiveFrom time.Time  `json:"effective_from"`
	EffectiveTo   *time.Time `json:"effective_to,omitempty"`
	IsActive      bool       `json:"is_active"`
	Description   string     `json:"description"`
	BusinessId    uint       `json:"business_id"`
	BranchId      uint       `json:"branch_id"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type MenuMonthReq struct {
	Month   int    `json:"month" binding:"required"`
	Year    int    `json:"year" binding:"required"`
	Name    string `json:"name" binding:"required"`
	Status  string `json:"status"`
	Version int    `json:"version"`
	Notes   string `json:"notes"`
}

type MenuDayItemReq struct {
	MealSlot       string `json:"meal_slot" binding:"required"`
	ItemName       string `json:"item_name" binding:"required"`
	Classification string `json:"classification" binding:"required"`
	DisplayOrder   int    `json:"display_order"`
}

type MenuDayReq struct {
	MenuDate string           `json:"menu_date" binding:"required"`
	Items    []MenuDayItemReq `json:"items" binding:"required"`
}

type MenuMonthRes struct {
	ID         uint      `json:"id"`
	Month      int       `json:"month"`
	Year       int       `json:"year"`
	Name       string    `json:"name"`
	Status     string    `json:"status"`
	Version    int       `json:"version"`
	Notes      string    `json:"notes"`
	BusinessId uint      `json:"business_id"`
	BranchId   uint      `json:"branch_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type CustomerSubscriptionReq struct {
	CustomerID         uint   `json:"customer_id" binding:"required"`
	SubscriptionPlanID uint   `json:"subscription_plan_id" binding:"required"`
	StartDate          string `json:"start_date" binding:"required"`
	EndDate            string `json:"end_date" binding:"required"`
	MealPreference     string `json:"meal_preference" binding:"required"`
	Status             string `json:"status"`
	BillingStatus      string `json:"billing_status"`
	Notes              string `json:"notes"`
}

type CustomerSubscriptionRes struct {
	ID                 uint      `json:"id"`
	CustomerID         uint      `json:"customer_id"`
	SubscriptionPlanID uint      `json:"subscription_plan_id"`
	StartDate          time.Time `json:"start_date"`
	EndDate            time.Time `json:"end_date"`
	MealPreference     string    `json:"meal_preference"`
	Status             string    `json:"status"`
	BillingStatus      string    `json:"billing_status"`
	Notes              string    `json:"notes"`
	BusinessId         uint      `json:"business_id"`
	BranchId           uint      `json:"branch_id"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type CustomerSubscriptionDayRes struct {
	ID                     uint      `json:"id"`
	CustomerSubscriptionID uint      `json:"customer_subscription_id"`
	CustomerID             uint      `json:"customer_id"`
	ScheduleDate           time.Time `json:"schedule_date"`
	DayName                string    `json:"day_name"`
	BreakfastRequired      bool      `json:"breakfast_required"`
	LunchRequired          bool      `json:"lunch_required"`
	DinnerRequired         bool      `json:"dinner_required"`
	BreakfastSkipped       bool      `json:"breakfast_skipped"`
	LunchSkipped           bool      `json:"lunch_skipped"`
	DinnerSkipped          bool      `json:"dinner_skipped"`
	MealPreference         string    `json:"meal_preference"`
	MenuDayID              *uint     `json:"menu_day_id,omitempty"`
	Status                 string    `json:"status"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}
