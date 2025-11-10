package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (user *User) BeforeCreate(tx *gorm.DB) (err error) {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	return
}

func (profile *StudentProfile) BeforeCreate(tx *gorm.DB) (err error) {
	if profile.ID == uuid.Nil {
		profile.ID = uuid.New()
	}
	return
}

func (profile *InstructorProfile) BeforeCreate(tx *gorm.DB) (err error) {
	if profile.ID == uuid.Nil {
		profile.ID = uuid.New()
	}
	return
}

func (profile *AdminProfile) BeforeCreate(tx *gorm.DB) (err error) {
	if profile.ID == uuid.Nil {
		profile.ID = uuid.New()
	}
	return
}

// --- MODELS ---

type User struct {
	ID                uuid.UUID `gorm:"type:uuid;primary_key"`
	Email             string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	PasswordHash      string    `gorm:"type:varchar(255);not null"`
	Role              string    `gorm:"type:varchar(50);not null;index"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
	StudentProfile    *StudentProfile    `gorm:"foreignKey:UserID"`
	InstructorProfile *InstructorProfile `gorm:"foreignKey:UserID"`
	AdminProfile      *AdminProfile      `gorm:"foreignKey:UserID"`
}

type StudentProfile struct {
	ID              uuid.UUID `gorm:"type:uuid;primary_key"`
	UserID          uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	FullName        string    `gorm:"type:varchar(255);not null"`
	RollNo          string    `gorm:"type:varchar(50);not null;uniqueIndex:uni_student_profiles_roll_no"`
	Branch          string    `gorm:"type:varchar(100)"`
	YearOfAdmission *int      `gorm:""`
	IsTA            bool      `gorm:"default:false"`

	// --- ENHANCED PERSONAL DETAIL FIELDS ---
	DateOfBirth   *time.Time `gorm:""`
	Address       string     `gorm:"type:text"`
	ContactNumber string     `gorm:"type:varchar(20)"`
	FatherName    string     `gorm:"type:varchar(255)"`
	MotherName    string     `gorm:"type:varchar(255)"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

// ... (InstructorProfile and AdminProfile remain the same) ...

type InstructorProfile struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	FullName   string    `gorm:"type:varchar(255);not null"`
	EmployeeID string    `gorm:"type:varchar(50);not null;uniqueIndex:uni_instructor_profiles_employee_id"`
	Department string    `gorm:"type:varchar(100)"`
	Title      string    `gorm:"type:varchar(100)"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type AdminProfile struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	FullName   string    `gorm:"type:varchar(255);not null"`
	EmployeeID string    `gorm:"type:varchar(50);not null;uniqueIndex:uni_admin_profiles_employee_id"`
	JobTitle   string    `gorm:"type:varchar(100)"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
