package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/yogalink/yogalink-api/internal/model"
	"github.com/yogalink/yogalink-api/internal/service"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// Error 错误响应
func Error(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

// JWTClaims JWT声明
type JWTClaims struct {
	UserID uint64 `json:"user_id"`
	jwt.RegisteredClaims
}

var jwtSecret = []byte("your-secret-key-change-in-production")

// JWTAuth JWT认证中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			Error(c, 401, "未登录")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			Error(c, 401, "认证格式错误")
			c.Abort()
			return
		}

		claims := &JWTClaims{}
		token, err := jwt.ParseWithClaims(parts[1], claims, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			Error(c, 401, "登录已过期")
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Next()
	}
}

// GenerateToken 生成JWT令牌
func GenerateToken(userID uint64) (string, error) {
	claims := &JWTClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// UserHandler 用户处理器
type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// Register 用户注册
func (h *UserHandler) Register(c *gin.Context) {
	var req service.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "参数错误: "+err.Error())
		return
	}

	user, err := h.userService.Register(c.Request.Context(), &req)
	if err != nil {
		Error(c, 400, err.Error())
		return
	}

	token, _ := GenerateToken(user.ID)
	Success(c, gin.H{
		"user":  toUserResponse(user),
		"token": token,
	})
}

// Login 用户登录
func (h *UserHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "参数错误")
		return
	}

	user, err := h.userService.Login(c.Request.Context(), &req)
	if err != nil {
		Error(c, 400, err.Error())
		return
	}

	token, _ := GenerateToken(user.ID)
	Success(c, gin.H{
		"user":  toUserResponse(user),
		"token": token,
	})
}

// GetProfile 获取用户信息
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := c.GetUint64("userID")
	user, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		Error(c, 404, "用户不存在")
		return
	}

	Success(c, toUserResponse(user))
}

func toUserResponse(user *model.User) gin.H {
	return gin.H{
		"id":         user.ID,
		"uuid":       user.UUID,
		"phone":      user.Phone,
		"nickname":   user.Nickname,
		"avatar":     user.Avatar,
		"gender":     user.Gender,
		"created_at": user.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// CourseHandler 课程处理器
type CourseHandler struct {
	courseService *service.CourseService
}

func NewCourseHandler(courseService *service.CourseService) *CourseHandler {
	return &CourseHandler{courseService: courseService}
}

// ListSchedules 获取课程列表
func (h *CourseHandler) ListSchedules(c *gin.Context) {
	date := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.courseService.ListSchedules(c.Request.Context(), date, page, pageSize)
	if err != nil {
		Error(c, 500, "获取课程列表失败")
		return
	}

	Success(c, result)
}

// GetScheduleDetail 获取课程详情
func (h *CourseHandler) GetScheduleDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "参数错误")
		return
	}

	schedule, err := h.courseService.GetScheduleDetail(c.Request.Context(), id)
	if err != nil {
		Error(c, 404, "课程不存在")
		return
	}

	Success(c, schedule)
}

// BookingHandler 预约处理器
type BookingHandler struct {
	bookingService *service.BookingService
}

func NewBookingHandler(bookingService *service.BookingService) *BookingHandler {
	return &BookingHandler{bookingService: bookingService}
}

// CreateBooking 创建预约
func (h *BookingHandler) CreateBooking(c *gin.Context) {
	userID := c.GetUint64("userID")

	var req service.CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "参数错误")
		return
	}

	req.UserID = userID
	booking, err := h.bookingService.CreateBooking(c.Request.Context(), &req)
	if err != nil {
		Error(c, 400, err.Error())
		return
	}

	Success(c, booking)
}

// GetBookingDetail 获取预约详情
func (h *BookingHandler) GetBookingDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "参数错误")
		return
	}

	booking, err := h.bookingService.GetBookingDetail(c.Request.Context(), id)
	if err != nil {
		Error(c, 404, "预约不存在")
		return
	}

	// 检查权限
	userID := c.GetUint64("userID")
	if booking.UserID != userID {
		Error(c, 403, "无权查看")
		return
	}

	Success(c, booking)
}

// ListMyBookings 获取我的预约列表
func (h *BookingHandler) ListMyBookings(c *gin.Context) {
	userID := c.GetUint64("userID")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.bookingService.ListUserBookings(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		Error(c, 500, "获取预约列表失败")
		return
	}

	Success(c, result)
}

// CancelBooking 取消预约
func (h *BookingHandler) CancelBooking(c *gin.Context) {
	bookingID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "参数错误")
		return
	}

	userID := c.GetUint64("userID")
	if err := h.bookingService.CancelBooking(c.Request.Context(), bookingID, userID); err != nil {
		Error(c, 400, err.Error())
		return
	}

	Success(c, nil)
}

// StudioHandler 场馆处理器
type StudioHandler struct {
	studioService *service.StudioService
}

func NewStudioHandler(studioService *service.StudioService) *StudioHandler {
	return &StudioHandler{studioService: studioService}
}

// ListStudios 获取场馆列表
func (h *StudioHandler) ListStudios(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.studioService.ListStudios(c.Request.Context(), page, pageSize)
	if err != nil {
		Error(c, 500, "获取场馆列表失败")
		return
	}

	Success(c, result)
}

// GetStudioDetail 获取场馆详情
func (h *StudioHandler) GetStudioDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "参数错误")
		return
	}

	studio, err := h.studioService.GetStudioDetail(c.Request.Context(), id)
	if err != nil {
		Error(c, 404, "场馆不存在")
		return
	}

	Success(c, studio)
}
