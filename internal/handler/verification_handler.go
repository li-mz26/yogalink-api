package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yogalink/yogalink-api/internal/model"
	"github.com/yogalink/yogalink-api/internal/service"
)

// VerificationHandler 实名认证处理器
type VerificationHandler struct {
	verificationService *service.VerificationService
}

func NewVerificationHandler(verificationService *service.VerificationService) *VerificationHandler {
	return &VerificationHandler{verificationService: verificationService}
}

// SubmitRealNameVerification 提交实名认证
func (h *VerificationHandler) SubmitRealNameVerification(c *gin.Context) {
	userID := c.GetUint64("userID")

	var req service.RealNameVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "参数错误: "+err.Error())
		return
	}

	req.UserID = userID
	verification, err := h.verificationService.SubmitVerification(c.Request.Context(), &req)
	if err != nil {
		Error(c, 400, err.Error())
		return
	}

	Success(c, verification)
}

// GetVerificationStatus 获取实名认证状态
func (h *VerificationHandler) GetVerificationStatus(c *gin.Context) {
	userID := c.GetUint64("userID")

	status, err := h.verificationService.GetVerificationStatus(c.Request.Context(), userID)
	if err != nil {
		Error(c, 500, "获取状态失败")
		return
	}

	Success(c, status)
}

// FaceVerification 上课前人脸识别
func (h *VerificationHandler) FaceVerification(c *gin.Context) {
	userID := c.GetUint64("userID")

	bookingID, err := strconv.ParseUint(c.Param("booking_id"), 10, 64)
	if err != nil {
		Error(c, 400, "参数错误")
		return
	}

	var req struct {
		FaceImage string `json:"face_image" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "参数错误: 需要提供人脸照片")
		return
	}

	result, err := h.verificationService.FaceVerification(c.Request.Context(), userID, bookingID, req.FaceImage)
	if err != nil {
		Error(c, 400, err.Error())
		return
	}

	Success(c, result)
}

// CheckFaceVerificationStatus 检查人脸识别状态
func (h *VerificationHandler) CheckFaceVerificationStatus(c *gin.Context) {
	userID := c.GetUint64("userID")

	bookingID, err := strconv.ParseUint(c.Param("booking_id"), 10, 64)
	if err != nil {
		Error(c, 400, "参数错误")
		return
	}

	status, err := h.verificationService.CheckFaceVerificationStatus(c.Request.Context(), userID, bookingID)
	if err != nil {
		Error(c, 500, err.Error())
		return
	}

	Success(c, status)
}

// UploadIDCardImage 上传身份证照片（获取临时URL）
func (h *VerificationHandler) UploadIDCardImage(c *gin.Context) {
	// 这里处理身份证照片上传
	// 实际生产环境应该使用云存储（如阿里云OSS、腾讯云COS等）
	Success(c, gin.H{
		"url": "https://example.com/temp/image.jpg",
		"note": "请接入实际的对象存储服务",
	})
}

// VerificationMiddleware 实名认证检查中间件
func VerificationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint64("userID")

		// 跳过某些路径
		path := c.Request.URL.Path
		if path == "/v1/user/verification" || path == "/v1/user/verification/status" {
			c.Next()
			return
		}

		// 获取用户实名认证状态
		// TODO: 从缓存或数据库获取
		// 这里简化处理，实际需要根据业务逻辑判断

		_ = userID
		c.Next()
	}
}
