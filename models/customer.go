package models

import "time"

type Customer struct {
	ID                    uint       `json:"id" gorm:"primaryKey"`
	CustomerCode          string     `json:"customer_code" gorm:"size:50;not null;uniqueIndex"`
	FullName              string     `json:"full_name" gorm:"size:150;not null"`
	PrimaryPhone          string     `json:"primary_phone" gorm:"size:30;not null;index"`
	AlternatePhone        string     `json:"alternate_phone" gorm:"size:30"`
	Email                 string     `json:"email" gorm:"size:150;index"`
	Password              string     `json:"-" gorm:"size:255"`
	Gender                string     `json:"gender" gorm:"type:varchar(20);check:gender IN ('male','female','other','unknown')"`
	DefaultMealPreference string     `json:"default_meal_preference" gorm:"type:varchar(20);not null;default:'veg';check:default_meal_preference IN ('veg','egg','non_veg')"`
	Status                string     `json:"status" gorm:"type:varchar(20);not null;default:'active';check:status IN ('active','inactive')"`
	Notes                 string     `json:"notes"`
	BusinessId            uint       `json:"business_id" gorm:"index;not null"`
	BranchId              uint       `json:"branch_id" gorm:"index;not null"`
	CreatedBy             uint       `json:"created_by"`
	UpdatedBy             uint       `json:"updated_by"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
	DeletedAt             *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

type CustomerAddress struct {
	ID             uint       `json:"id" gorm:"primaryKey"`
	CustomerID     uint       `json:"customer_id" gorm:"index;not null"`
	AddressType    string     `json:"address_type" gorm:"type:varchar(20);not null;check:address_type IN ('default','breakfast','lunch','dinner')"`
	AddressLine    string     `json:"address_line" gorm:"not null"`
	Area           string     `json:"area" gorm:"size:120;index"`
	Landmark       string     `json:"landmark" gorm:"size:255"`
	ContactPerson  string     `json:"contact_person" gorm:"size:120"`
	AlternatePhone string     `json:"alternate_phone" gorm:"size:30"`
	Latitude       string     `json:"latitude" gorm:"size:50"`
	Longitude      string     `json:"longitude" gorm:"size:50"`
	DeliveryNotes  string     `json:"delivery_notes"`
	IsPrimary      bool       `json:"is_primary" gorm:"default:false"`
	BusinessId     uint       `json:"business_id" gorm:"index;not null"`
	BranchId       uint       `json:"branch_id" gorm:"index;not null"`
	CreatedBy      uint       `json:"created_by"`
	UpdatedBy      uint       `json:"updated_by"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}
