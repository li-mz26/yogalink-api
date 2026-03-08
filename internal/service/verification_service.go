package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/yogalink/yogalink-api/internal/model"
	"github.com/yogalink/yogalink-api/internal/repository"
)

// VerificationService 实名认证服务
type VerificationService struct {
	verificationRepo *repository.VerificationRepository
	userRepo         *repository.UserRepository
}

func NewVerificationService(
	verificationRepo *repository.VerificationRepository,
	userRepo *repository.UserRepository,
) *VerificationService {
	return &VerificationService{
		verificationRepo: verificationRepo,
		userRepo:         userRepo,
	}
}

// RealNameVerificationRequest 实名认证请求
type RealNameVerificationRequest struct {
	UserID       uint64 `json:"user_id"`
	RealName     string `json:"real_name" binding:"required"`
	IDCardNumber string `json:"id_card_number" binding:"required,len=18"`
	IDCardFront  string `json:"id_card_front" binding:"required,url"`
	IDCardBack   string `json:"id_card_back" binding:"required,url"`
}

// SubmitVerification 提交实名认证
func (s *VerificationService) SubmitVerification(ctx context.Context, req *RealNameVerificationRequest) (*model.RealNameVerification, error) {
	// 解析身份证信息
	idInfo, err := s.parseIDCard(req.IDCardNumber)
	if err != nil {
		return nil, err
	}

	// 验证是否为女性
	if idInfo.Gender != 2 {
		return nil, errors.New("本平台仅限女性用户")
	}

	// 验证是否成年人
	if !idInfo.IsAdult {
		return nil, errors.New("本平台仅限18岁以上用户")
	}

	// 创建或更新认证记录
	verification := &model.RealNameVerification{
		UserID:       req.UserID,
		RealName:     req.RealName,
		IDCardNumber: s.encryptIDCard(req.IDCardNumber),
		IDCardFront:  req.IDCardFront,
		IDCardBack:   req.IDCardBack,
		Gender:       idInfo.Gender,
		BirthDate:    idInfo.BirthDate,
		Status:       model.VerificationStatusPending,
	}

	if err := s.verificationRepo.CreateOrUpdate(ctx, verification); err != nil {
		return nil, err
	}

	// 更新用户认证状态为审核中
	if err := s.userRepo.UpdateVerificationStatus(ctx, req.UserID, model.VerificationStatusPending); err != nil {
		// 记录日志
	}

	return verification, nil
}

// GetVerificationStatus 获取认证状态
func (s *VerificationService) GetVerificationStatus(ctx context.Context, userID uint64) (gin.H, error) {
	verification, err := s.verificationRepo.GetByUserID(ctx, userID)
	if err != nil {
		return gin.H{
			"status": model.VerificationStatusUnverified,
		}, nil
	}

	return gin.H{
		"status":       verification.Status,
		"real_name":    verification.RealName,
		"gender":       verification.Gender,
		"birth_date":   verification.BirthDate,
		"is_adult":     verification.IsAdult(),
		"reject_reason": verification.RejectReason,
		"verified_at":  verification.VerifiedAt,
	}, nil
}

// FaceVerificationResult 人脸识别结果
type FaceVerificationResult struct {
	IsMatched  bool    `json:"is_matched"`
	MatchScore float64 `json:"match_score"`
	Message    string  `json:"message"`
}

// FaceVerification 人脸识别验证
func (s *VerificationService) FaceVerification(ctx context.Context, userID, bookingID uint64, faceImage string) (*FaceVerificationResult, error) {
	// 1. 检查用户是否已实名认证
	verification, err := s.verificationRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, errors.New("请先完成实名认证")
	}

	if verification.Status != model.VerificationStatusVerified {
		return nil, errors.New("实名认证未通过，无法使用人脸识别")
	}

	// 2. 检查预约是否有效
	// TODO: 检查预约是否存在且状态为 accepted

	// 3. 调用第三方人脸识别API进行比对
	// TODO: 接入实际的人脸识别服务（如阿里云、腾讯云、百度AI等）
	// 这里使用模拟数据
	matchScore := s.simulateFaceMatch()
	isMatched := matchScore >= 0.85 // 85%相似度阈值

	// 4. 记录验证结果
	record := &model.FaceVerificationRecord{
		UserID:     userID,
		BookingID:  bookingID,
		VerifyType: "pre_class",
		FaceImage:  faceImage,
		MatchScore: matchScore,
		IsMatched:  isMatched,
		VerifiedAt: time.Now(),
	}

	if err := s.verificationRepo.CreateFaceVerificationRecord(ctx, record); err != nil {
		// 记录日志但不影响返回
	}

	if !isMatched {
		return &FaceVerificationResult{
			IsMatched:  false,
			MatchScore: matchScore,
			Message:    "人脸识别未通过，请重试",
		}, nil
	}

	return &FaceVerificationResult{
		IsMatched:  true,
		MatchScore: matchScore,
		Message:    "人脸识别通过",
	}, nil
}

// CheckFaceVerificationStatus 检查人脸识别状态
func (s *VerificationService) CheckFaceVerificationStatus(ctx context.Context, userID, bookingID uint64) (gin.H, error) {
	record, err := s.verificationRepo.GetLatestFaceVerification(ctx, userID, bookingID)
	if err != nil {
		return gin.H{
			"is_verified": false,
			"message":     "尚未进行人脸识别",
		}, nil
	}

	return gin.H{
		"is_verified": record.IsMatched,
		"match_score": record.MatchScore,
		"verified_at": record.VerifiedAt,
	}, nil
}

// parseIDCard 解析身份证信息
func (s *VerificationService) parseIDCard(idCard string) (*IDCardInfo, error) {
	if len(idCard) != 18 {
		return nil, errors.New("身份证号格式不正确")
	}

	// 解析出生日期
	birthDate := fmt.Sprintf("%s-%s-%s", idCard[6:10], idCard[10:12], idCard[12:14])

	// 解析性别（倒数第二位，奇数为男，偶数为女）
	genderCode := idCard[16]
	gender := 1 // 男
	if (genderCode-'0')%2 == 0 {
		gender = 2 // 女
	}

	// 计算年龄
	birth, _ := time.Parse("2006-01-02", birthDate)
	isAdult := time.Since(birth).Hours()/24/365 >= 18

	return &IDCardInfo{
		BirthDate: birthDate,
		Gender:    gender,
		IsAdult:   isAdult,
	}, nil
}

// IDCardInfo 身份证信息
type IDCardInfo struct {
	BirthDate string
	Gender    int8
	IsAdult   bool
}

// encryptIDCard 加密身份证号（简单示例，实际应使用更强的加密）
func (s *VerificationService) encryptIDCard(idCard string) string {
	// 保留前3位和后4位，中间加密
	if len(idCard) < 8 {
		return idCard
	}
	return idCard[:3] + "****" + idCard[len(idCard)-4:]
}

// simulateFaceMatch 模拟人脸识别匹配
func (s *VerificationService) simulateFaceMatch() float64 {
	// 实际应该调用第三方API
	// 返回 0.8-0.99 之间的随机数模拟
	return 0.8 + float64(time.Now().UnixNano()%20)/100
}

// ApproveVerification 管理员通过认证（内部使用）
func (s *VerificationService) ApproveVerification(ctx context.Context, userID uint64) error {
	now := time.Now()
	if err := s.verificationRepo.UpdateStatus(ctx, userID, model.VerificationStatusVerified, "", &now); err != nil {
		return err
	}
	return s.userRepo.UpdateVerificationStatus(ctx, userID, model.VerificationStatusVerified)
}

// RejectVerification 管理员拒绝认证（内部使用）
func (s *VerificationService) RejectVerification(ctx context.Context, userID uint64, reason string) error {
	if err := s.verificationRepo.UpdateStatus(ctx, userID, model.VerificationStatusRejected, reason, nil); err != nil {
		return err
	}
	return s.userRepo.UpdateVerificationStatus(ctx, userID, model.VerificationStatusRejected)
}
