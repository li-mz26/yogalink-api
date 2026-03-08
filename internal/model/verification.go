package model

import (
	"time"
)

// 实名认证状态
type VerificationStatus string

const (
	VerificationStatusUnverified VerificationStatus = "unverified" // 未认证
	VerificationStatusPending    VerificationStatus = "pending"    // 审核中
	VerificationStatusVerified   VerificationStatus = "verified"   // 已认证
	VerificationStatusRejected   VerificationStatus = "rejected"   // 认证失败
)

// RealNameVerification 实名认证信息
type RealNameVerification struct {
	ID              uint64             `gorm:"primarykey" json:"id"`
	UserID          uint64             `gorm:"uniqueIndex" json:"user_id"`
	RealName        string             `gorm:"size:50" json:"real_name"`           // 真实姓名
	IDCardNumber    string             `gorm:"size:18" json:"id_card_number"`      // 身份证号（加密存储）
	IDCardFront     string             `gorm:"size:500" json:"id_card_front"`      // 身份证正面照URL
	IDCardBack      string             `gorm:"size:500" json:"id_card_back"`       // 身份证背面照URL
	Gender          int8               `json:"gender"`                              // 1-男 2-女
	BirthDate       string             `gorm:"size:10" json:"birth_date"`          // 出生日期 YYYY-MM-DD
	Status          VerificationStatus `gorm:"size:20;default:'unverified'" json:"status"`
	RejectReason    string             `json:"reject_reason,omitempty"`             // 拒绝原因
	VerifiedAt      *time.Time         `json:"verified_at,omitempty"`               // 认证通过时间
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
}

// FaceVerificationRecord 人脸识别记录
type FaceVerificationRecord struct {
	ID           uint64    `gorm:"primarykey" json:"id"`
	UserID       uint64    `gorm:"index" json:"user_id"`
	BookingID    uint64    `gorm:"index" json:"booking_id"`           // 关联的预约
	VerifyType   string    `gorm:"size:20" json:"verify_type"`         // 类型：pre_class（上课前）
	FaceImage    string    `gorm:"size:500" json:"face_image"`         // 人脸照片URL
	MatchScore   float64   `json:"match_score"`                        // 匹配分数
	IsMatched    bool      `json:"is_matched"`                         // 是否匹配
	VerifiedAt   time.Time `json:"verified_at"`
	CreatedAt    time.Time `json:"created_at"`
}

// IsAdult 检查是否成年人（18岁以上）
func (v *RealNameVerification) IsAdult() bool {
	if v.BirthDate == "" {
		return false
	}
	birth, err := time.Parse("2006-01-02", v.BirthDate)
	if err != nil {
		return false
	}
	return time.Since(birth).Hours() / 24 / 365 >= 18
}

// IsFemale 检查是否为女性
func (v *RealNameVerification) IsFemale() bool {
	return v.Gender == 2
}

// CanUseApp 检查是否可以使用App（已认证的女性成年人）
func (v *RealNameVerification) CanUseApp() bool {
	return v.Status == VerificationStatusVerified && v.IsFemale() && v.IsAdult()
}
