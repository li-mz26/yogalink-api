package repository

import (
	"context"
	"time"

	"github.com/yogalink/yogalink-api/internal/model"
	"gorm.io/gorm"
)

// VerificationRepository 实名认证数据访问
type VerificationRepository struct {
	db *gorm.DB
}

func NewVerificationRepository(db *gorm.DB) *VerificationRepository {
	return &VerificationRepository{db: db}
}

// CreateOrUpdate 创建或更新实名认证信息
func (r *VerificationRepository) CreateOrUpdate(ctx context.Context, verification *model.RealNameVerification) error {
	var existing model.RealNameVerification
	err := r.db.WithContext(ctx).Where("user_id = ?", verification.UserID).First(&existing).Error
	if err == nil {
		// 存在则更新
		verification.ID = existing.ID
		return r.db.WithContext(ctx).Save(verification).Error
	}
	// 不存在则创建
	return r.db.WithContext(ctx).Create(verification).Error
}

// GetByUserID 根据用户ID获取认证信息
func (r *VerificationRepository) GetByUserID(ctx context.Context, userID uint64) (*model.RealNameVerification, error) {
	var verification model.RealNameVerification
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&verification).Error
	if err != nil {
		return nil, err
	}
	return &verification, nil
}

// UpdateStatus 更新认证状态
func (r *VerificationRepository) UpdateStatus(ctx context.Context, userID uint64, status model.VerificationStatus, rejectReason string, verifiedAt *time.Time) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if rejectReason != "" {
		updates["reject_reason"] = rejectReason
	}
	if verifiedAt != nil {
		updates["verified_at"] = verifiedAt
	}
	return r.db.WithContext(ctx).Model(&model.RealNameVerification{}).
		Where("user_id = ?", userID).
		Updates(updates).Error
}

// CreateFaceVerificationRecord 创建人脸识别记录
func (r *VerificationRepository) CreateFaceVerificationRecord(ctx context.Context, record *model.FaceVerificationRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

// GetLatestFaceVerification 获取最新的人脸识别记录
func (r *VerificationRepository) GetLatestFaceVerification(ctx context.Context, userID, bookingID uint64) (*model.FaceVerificationRecord, error) {
	var record model.FaceVerificationRecord
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND booking_id = ?", userID, bookingID).
		Order("created_at DESC").
		First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// ListPendingVerifications 获取待审核的认证列表（管理员用）
func (r *VerificationRepository) ListPendingVerifications(ctx context.Context, page, pageSize int) ([]model.RealNameVerification, int64, error) {
	var verifications []model.RealNameVerification
	var total int64

	query := r.db.WithContext(ctx).Model(&model.RealNameVerification{}).
		Where("status = ?", model.VerificationStatusPending)

	query.Count(&total)
	err := query.Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&verifications).Error

	return verifications, total, err
}
