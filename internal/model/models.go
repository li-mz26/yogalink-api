package model

import (
	"time"
)

// User 用户模型
type User struct {
	ID        uint64    `gorm:"primarykey" json:"id"`
	UUID      string    `gorm:"uniqueIndex;size:36" json:"uuid"`
	Phone     string    `gorm:"uniqueIndex;size:20" json:"phone"`
	Email     string    `gorm:"uniqueIndex;size:100" json:"email,omitempty"`
	Password  string    `gorm:"size:255" json:"-"`
	Nickname  string    `gorm:"size:50" json:"nickname"`
	Avatar    string    `gorm:"size:500" json:"avatar"`
	Gender    int8      `gorm:"default:0" json:"gender"` // 0-未知 1-男 2-女
	Status    int8      `gorm:"default:1" json:"status"` // 0-禁用 1-正常
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserProfile 用户瑜伽档案
type UserProfile struct {
	ID               uint64   `gorm:"primarykey" json:"id"`
	UserID           uint64   `gorm:"uniqueIndex" json:"user_id"`
	ExperienceLevel  int8     `json:"experience_level"` // 1-新手 2-进阶 3-高手
	Goals            []string `gorm:"serializer:json" json:"goals"`
	PreferredStyles  []string `gorm:"serializer:json" json:"preferred_styles"`
	PhysicalCondition string  `json:"physical_condition,omitempty"`
}

// Studio 瑜伽场馆
type Studio struct {
	ID          uint64    `gorm:"primarykey" json:"id"`
	Name        string    `gorm:"size:100" json:"name"`
	Description string    `json:"description"`
	CoverImage  string    `gorm:"size:500" json:"cover_image"`
	Address     string    `gorm:"size:255" json:"address"`
	Longitude   float64   `json:"longitude"`
	Latitude    float64   `json:"latitude"`
	Phone       string    `gorm:"size:20" json:"phone"`
	Rating      float32   `gorm:"default:5.0" json:"rating"`
	Status      int8      `gorm:"default:1" json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

// Instructor 教练
type Instructor struct {
	ID            uint64   `gorm:"primarykey" json:"id"`
	StudioID      uint64   `json:"studio_id"`
	Name          string   `gorm:"size:50" json:"name"`
	Avatar        string   `gorm:"size:500" json:"avatar"`
	Bio           string   `json:"bio"`
	Specialties   []string `gorm:"serializer:json" json:"specialties"`
	TeachingYears int      `json:"teaching_years"`
	Rating        float32  `gorm:"default:5.0" json:"rating"`
}

// CourseTemplate 课程模板
type CourseTemplate struct {
	ID          uint64   `gorm:"primarykey" json:"id"`
	StudioID    uint64   `json:"studio_id"`
	InstructorID uint64  `json:"instructor_id"`
	Name        string   `gorm:"size:100" json:"name"`
	Style       string   `gorm:"size:50" json:"style"` // 流派：哈他/流瑜伽/阴瑜伽等
	Level       int8     `json:"level"` // 1-初级 2-中级 3-高级
	Duration    int      `json:"duration"` // 分钟
	MaxStudents int      `json:"max_students"`
	Description string   `json:"description"`
	CoverImage  string   `gorm:"size:500" json:"cover_image"`
	Price       float64  `json:"price"`
}

// CourseSchedule 课程排期
type CourseSchedule struct {
	ID           uint64    `gorm:"primarykey" json:"id"`
	TemplateID   uint64    `json:"template_id"`
	StudioID     uint64    `json:"studio_id"`
	InstructorID uint64    `json:"instructor_id"`
	CourseDate   string    `gorm:"size:10" json:"course_date"` // YYYY-MM-DD
	StartTime    string    `gorm:"size:5" json:"start_time"`   // HH:MM
	EndTime      string    `gorm:"size:5" json:"end_time"`
	MaxStudents  int       `json:"max_students"`
	BookedCount  int       `gorm:"default:0" json:"booked_count"`
	Price        float64   `json:"price"`
	Status       int8      `gorm:"default:1" json:"status"` // 0-取消 1-正常 2-结束
	CreatedAt    time.Time `json:"created_at"`
	// 关联数据
	Template   CourseTemplate `gorm:"foreignKey:TemplateID" json:"template,omitempty"`
	Studio     Studio         `gorm:"foreignKey:StudioID" json:"studio,omitempty"`
	Instructor Instructor     `gorm:"foreignKey:InstructorID" json:"instructor,omitempty"`
}

// Booking 预约订单
type Booking struct {
	ID            uint64    `gorm:"primarykey" json:"id"`
	BookingNo     string    `gorm:"uniqueIndex;size:32" json:"booking_no"`
	UserID        uint64    `json:"user_id"`
	ScheduleID    uint64    `json:"schedule_id"`
	Status        int8      `json:"status"` // 0-取消 1-待支付 2-已支付 3-已完成 4-已评价
	Price         float64   `json:"price"`
	CheckinCode   string    `gorm:"size:20" json:"checkin_code,omitempty"`
	CheckinStatus int8      `gorm:"default:0" json:"checkin_status"` // 0-未签到 1-已签到
	PaidAt        *time.Time `json:"paid_at,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	// 关联数据
	Schedule CourseSchedule `gorm:"foreignKey:ScheduleID" json:"schedule,omitempty"`
}
