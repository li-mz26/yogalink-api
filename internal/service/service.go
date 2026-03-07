package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yogalink/yogalink-api/internal/model"
	"github.com/yogalink/yogalink-api/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// UserService 用户服务
type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required,min:6"`
	Nickname string `json:"nickname"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UserResponse 用户响应
type UserResponse struct {
	ID        uint64 `json:"id"`
	UUID      string `json:"uuid"`
	Phone     string `json:"phone"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
	Gender    int8   `json:"gender"`
	CreatedAt string `json:"created_at"`
}

// Register 用户注册
func (s *UserService) Register(ctx context.Context, req *RegisterRequest) (*model.User, error) {
	// 检查手机号是否已注册
	_, err := s.userRepo.GetByPhone(ctx, req.Phone)
	if err == nil {
		return nil, errors.New("手机号已注册")
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 创建用户
	user := &model.User{
		UUID:     uuid.New().String(),
		Phone:    req.Phone,
		Password: string(hashedPassword),
		Nickname: req.Nickname,
		Status:   1,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login 用户登录
func (s *UserService) Login(ctx context.Context, req *LoginRequest) (*model.User, error) {
	user, err := s.userRepo.GetByPhone(ctx, req.Phone)
	if err != nil {
		return nil, errors.New("手机号或密码错误")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("手机号或密码错误")
	}

	return user, nil
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(ctx context.Context, id uint64) (*model.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

// CourseService 课程服务
type CourseService struct {
	courseRepo  *repository.CourseRepository
	bookingRepo *repository.BookingRepository
}

func NewCourseService(courseRepo *repository.CourseRepository, bookingRepo *repository.BookingRepository) *CourseService {
	return &CourseService{
		courseRepo:  courseRepo,
		bookingRepo: bookingRepo,
	}
}

// ScheduleListResponse 课程列表响应
type ScheduleListResponse struct {
	List       []model.CourseSchedule `json:"list"`
	Total      int64                  `json:"total"`
	Page       int                    `json:"page"`
	PageSize   int                    `json:"page_size"`
	TotalPages int                    `json:"total_pages"`
}

// ListSchedules 获取课程列表
func (s *CourseService) ListSchedules(ctx context.Context, date string, page, pageSize int) (*ScheduleListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	schedules, total, err := s.courseRepo.ListSchedules(ctx, date, page, pageSize)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &ScheduleListResponse{
		List:       schedules,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetScheduleDetail 获取课程详情
func (s *CourseService) GetScheduleDetail(ctx context.Context, id uint64) (*model.CourseSchedule, error) {
	return s.courseRepo.GetScheduleByID(ctx, id)
}

// BookingService 预约服务
type BookingService struct {
	bookingRepo *repository.BookingRepository
	courseRepo  *repository.CourseRepository
}

func NewBookingService(bookingRepo *repository.BookingRepository, courseRepo *repository.CourseRepository) *BookingService {
	return &BookingService{
		bookingRepo: bookingRepo,
		courseRepo:  courseRepo,
	}
}

// CreateBookingRequest 创建预约请求
type CreateBookingRequest struct {
	UserID     uint64 `json:"user_id" binding:"required"`
	ScheduleID uint64 `json:"schedule_id" binding:"required"`
}

// CreateBooking 创建预约
func (s *BookingService) CreateBooking(ctx context.Context, req *CreateBookingRequest) (*model.Booking, error) {
	// 获取课程排期
	schedule, err := s.courseRepo.GetScheduleByID(ctx, req.ScheduleID)
	if err != nil {
		return nil, errors.New("课程不存在")
	}

	// 检查是否已满
	if schedule.BookedCount >= schedule.MaxStudents {
		return nil, errors.New("课程已满")
	}

	// 检查课程是否已开始
	courseTime, _ := time.Parse("2006-01-02 15:04", schedule.CourseDate+" "+schedule.StartTime)
	if time.Now().After(courseTime) {
		return nil, errors.New("课程已开始")
	}

	// 生成预约单号
	bookingNo := fmt.Sprintf("BK%s%d", time.Now().Format("20060102"), time.Now().UnixNano()%1000000)

	// 创建预约
	booking := &model.Booking{
		BookingNo:  bookingNo,
		UserID:     req.UserID,
		ScheduleID: req.ScheduleID,
		Status:     2, // 直接设置为已支付（MVP简化）
		Price:      schedule.Price,
		CheckinCode: generateCheckinCode(),
		PaidAt:     &[]time.Time{time.Now()}[0],
	}

	if err := s.bookingRepo.Create(ctx, booking); err != nil {
		return nil, err
	}

	// 更新已预约人数
	if err := s.courseRepo.UpdateBookedCount(ctx, req.ScheduleID, 1); err != nil {
		// 记录日志，但不影响预约结果
	}

	return booking, nil
}

// GetBookingDetail 获取预约详情
func (s *BookingService) GetBookingDetail(ctx context.Context, id uint64) (*model.Booking, error) {
	return s.bookingRepo.GetByID(ctx, id)
}

// ListUserBookings 获取用户预约列表
func (s *BookingService) ListUserBookings(ctx context.Context, userID uint64, page, pageSize int) (*ScheduleListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	bookings, total, err := s.bookingRepo.ListByUser(ctx, userID, page, pageSize)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &ScheduleListResponse{
		List:       bookings,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// CancelBooking 取消预约
func (s *BookingService) CancelBooking(ctx context.Context, bookingID uint64, userID uint64) error {
	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return errors.New("预约不存在")
	}

	if booking.UserID != userID {
		return errors.New("无权操作")
	}

	if booking.Status == 0 {
		return errors.New("预约已取消")
	}

	// 检查取消时间（开课前12小时）
	schedule, err := s.courseRepo.GetScheduleByID(ctx, booking.ScheduleID)
	if err != nil {
		return err
	}

	courseTime, _ := time.Parse("2006-01-02 15:04", schedule.CourseDate+" "+schedule.StartTime)
	if time.Until(courseTime) < 12*time.Hour {
		return errors.New("开课12小时内不可取消")
	}

	// 更新预约状态
	booking.Status = 0
	if err := s.bookingRepo.Update(ctx, booking); err != nil {
		return err
	}

	// 释放名额
	if err := s.courseRepo.UpdateBookedCount(ctx, booking.ScheduleID, -1); err != nil {
		// 记录日志
	}

	return nil
}

// generateCheckinCode 生成6位签到码
func generateCheckinCode() string {
	return fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
}

// StudioService 场馆服务
type StudioService struct {
	studioRepo *repository.StudioRepository
}

func NewStudioService(studioRepo *repository.StudioRepository) *StudioService {
	return &StudioService{studioRepo: studioRepo}
}

// ListStudios 获取场馆列表
func (s *StudioService) ListStudios(ctx context.Context, page, pageSize int) (*ScheduleListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	studios, total, err := s.studioRepo.List(ctx, page, pageSize)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &ScheduleListResponse{
		List:       studios,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetStudioDetail 获取场馆详情
func (s *StudioService) GetStudioDetail(ctx context.Context, id uint64) (*model.Studio, error) {
	return s.studioRepo.GetByID(ctx, id)
}
