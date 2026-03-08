package model

import (
	"time"
)

// UserRole 用户角色类型
type UserRole string

const (
	UserRoleTeacher UserRole = "teacher"
	UserRoleStudent UserRole = "student"
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
	Role      UserRole  `gorm:"size:20;default:'student'" json:"role"` // teacher/student
	Status    int8      `gorm:"default:1" json:"status"` // 0-禁用 1-正常
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TeacherProfile 老师档案
type TeacherProfile struct {
	ID              uint64   `gorm:"primarykey" json:"id"`
	UserID          uint64   `gorm:"uniqueIndex" json:"user_id"`
	Bio             string   `json:"bio"`                           // 个人简介
	Specialties     []string `gorm:"serializer:json" json:"specialties"` // 专长流派
	TeachingYears   int      `json:"teaching_years"`                // 教龄
	HourlyRate      float64  `json:"hourly_rate"`                   // 时薪
	TeachingStyle   string   `json:"teaching_style"`                // 教学风格
	Certifications  []string `gorm:"serializer:json" json:"certifications"` // 资质证书
	Languages       []string `gorm:"serializer:json" json:"languages"`      // 授课语言
	IsVerified      bool     `gorm:"default:false" json:"is_verified"` // 是否认证
	Rating          float32  `gorm:"default:5.0" json:"rating"`     // 评分
	TotalReviews    int      `gorm:"default:0" json:"total_reviews"` // 评价数
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// TeacherAvailability 老师可用时间段
type TeacherAvailability struct {
	ID        uint64    `gorm:"primarykey" json:"id"`
	TeacherID uint64    `gorm:"index" json:"teacher_id"`
	Date      string    `gorm:"size:10;index" json:"date"`       // YYYY-MM-DD
	StartTime string    `gorm:"size:5" json:"start_time"`        // HH:MM
	EndTime   string    `gorm:"size:5" json:"end_time"`          // HH:MM
	IsBooked  bool      `gorm:"default:false" json:"is_booked"`  // 是否已被预约
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// LocationType 地点类型
type LocationType string

const (
	LocationTypeMetro   LocationType = "metro"   // 地铁站
	LocationTypeCustom  LocationType = "custom"  // 自定义位置
)

// TeachingLocation 老师授课地点
type TeachingLocation struct {
	ID          uint64       `gorm:"primarykey" json:"id"`
	TeacherID   uint64       `gorm:"index" json:"teacher_id"`
	Type        LocationType `gorm:"size:20" json:"type"`              // metro/custom
	Name        string       `gorm:"size:100" json:"name"`             // 地点名称（如"国贸站"或"朝阳区大望路"）
	MetroLine   string       `gorm:"size:50" json:"metro_line,omitempty"` // 地铁线路（如"1号线"）
	Address     string       `gorm:"size:255" json:"address"`          // 详细地址
	Longitude   float64      `json:"longitude"`                          // 经度
	Latitude    float64      `json:"latitude"`                           // 纬度
	RadiusKm    float64      `json:"radius_km"`                          // 服务范围（公里）
	IsDefault   bool         `gorm:"default:false" json:"is_default"`  // 是否为默认地点
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// BookingStatus 预约状态
type BookingStatus string

const (
	BookingStatusPending   BookingStatus = "pending"   // 待确认
	BookingStatusAccepted  BookingStatus = "accepted"  // 已接受
	BookingStatusRejected  BookingStatus = "rejected"  // 已拒绝
	BookingStatusCancelled BookingStatus = "cancelled" // 已取消
	BookingStatusCompleted BookingStatus = "completed" // 已完成
)

// BookingRequest 学生预约请求
type BookingRequest struct {
	ID            uint64        `gorm:"primarykey" json:"id"`
	RequestNo     string        `gorm:"uniqueIndex;size:32" json:"request_no"`
	StudentID     uint64        `gorm:"index" json:"student_id"`
	TeacherID     uint64        `gorm:"index" json:"teacher_id"`
	LocationID    uint64        `json:"location_id"`           // 选择的授课地点
	Date          string        `gorm:"size:10" json:"date"`   // YYYY-MM-DD
	StartTime     string        `gorm:"size:5" json:"start_time"` // HH:MM
	EndTime       string        `gorm:"size:5" json:"end_time"`   // HH:MM
	Duration      int           `json:"duration"`              // 课程时长（分钟）
	Status        BookingStatus `gorm:"size:20" json:"status"`
	Message       string        `json:"message"`               // 学生留言
	ResponseNote  string        `json:"response_note"`         // 老师回复
	Price         float64       `json:"price"`                 // 最终价格
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
	RespondedAt   *time.Time    `json:"responded_at,omitempty"`
	
	// 关联数据
	Student  User             `gorm:"foreignKey:StudentID" json:"student,omitempty"`
	Teacher  User             `gorm:"foreignKey:TeacherID" json:"teacher,omitempty"`
	Location TeachingLocation `gorm:"foreignKey:LocationID" json:"location,omitempty"`
}

// StudentProfile 学生档案（可选）
type StudentProfile struct {
	ID              uint64   `gorm:"primarykey" json:"id"`
	UserID          uint64   `gorm:"uniqueIndex" json:"user_id"`
	ExperienceLevel int8     `json:"experience_level"` // 1-新手 2-进阶 3-高手
	Goals           []string `gorm:"serializer:json" json:"goals"`
	PreferredStyles []string `gorm:"serializer:json" json:"preferred_styles"`
	HealthNotes     string   `json:"health_notes,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// ===== 以下保留原有模型用于兼容 =====

// Studio 瑜伽场馆（保留但不再使用）
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

// Instructor 教练（保留但不再使用）
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

// CourseTemplate 课程模板（保留但不再使用）
type CourseTemplate struct {
	ID          uint64   `gorm:"primarykey" json:"id"`
	StudioID    uint64   `json:"studio_id"`
	InstructorID uint64  `json:"instructor_id"`
	Name        string   `gorm:"size:100" json:"name"`
	Style       string   `gorm:"size:50" json:"style"`
	Level       int8     `json:"level"`
	Duration    int      `json:"duration"`
	MaxStudents int      `json:"max_students"`
	Description string   `json:"description"`
	CoverImage  string   `gorm:"size:500" json:"cover_image"`
	Price       float64  `json:"price"`
}

// CourseSchedule 课程排期（保留但不再使用）
type CourseSchedule struct {
	ID           uint64    `gorm:"primarykey" json:"id"`
	TemplateID   uint64    `json:"template_id"`
	StudioID     uint64    `json:"studio_id"`
	InstructorID uint64    `json:"instructor_id"`
	CourseDate   string    `gorm:"size:10" json:"course_date"`
	StartTime    string    `gorm:"size:5" json:"start_time"`
	EndTime      string    `gorm:"size:5" json:"end_time"`
	MaxStudents  int       `json:"max_students"`
	BookedCount  int       `gorm:"default:0" json:"booked_count"`
	Price        float64   `json:"price"`
	Status       int8      `gorm:"default:1" json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	Template     CourseTemplate `gorm:"foreignKey:TemplateID" json:"template,omitempty"`
	Studio       Studio         `gorm:"foreignKey:StudioID" json:"studio,omitempty"`
	Instructor   Instructor     `gorm:"foreignKey:InstructorID" json:"instructor,omitempty"`
}

// Booking 预约订单（保留但不再使用）
type Booking struct {
	ID            uint64     `gorm:"primarykey" json:"id"`
	BookingNo     string     `gorm:"uniqueIndex;size:32" json:"booking_no"`
	UserID        uint64     `json:"user_id"`
	ScheduleID    uint64     `json:"schedule_id"`
	Status        int8       `json:"status"`
	Price         float64    `json:"price"`
	CheckinCode   string     `gorm:"size:20" json:"checkin_code,omitempty"`
	CheckinStatus int8       `gorm:"default:0" json:"checkin_status"`
	PaidAt        *time.Time `json:"paid_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	Schedule      CourseSchedule `gorm:"foreignKey:ScheduleID" json:"schedule,omitempty"`
}
