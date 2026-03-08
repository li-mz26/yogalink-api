package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/yogalink/yogalink-api/internal/model"
	"github.com/yogalink/yogalink-api/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// ===== User Service =====

// UserService 用户服务
type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Phone    string          `json:"phone" binding:"required"`
	Password string          `json:"password" binding:"required,min:6"`
	Nickname string          `json:"nickname"`
	Role     model.UserRole  `json:"role" binding:"required,oneof=teacher student"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required"`
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
		Role:     req.Role,
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

// ===== Teacher Service =====

// TeacherService 老师服务
type TeacherService struct {
	profileRepo      *repository.TeacherProfileRepository
	availabilityRepo *repository.TeacherAvailabilityRepository
	locationRepo     *repository.TeachingLocationRepository
	bookingRepo      *repository.BookingRequestRepository
	userRepo         *repository.UserRepository
}

func NewTeacherService(
	profileRepo *repository.TeacherProfileRepository,
	availabilityRepo *repository.TeacherAvailabilityRepository,
	locationRepo *repository.TeachingLocationRepository,
	bookingRepo *repository.BookingRequestRepository,
	userRepo *repository.UserRepository,
) *TeacherService {
	return &TeacherService{
		profileRepo:      profileRepo,
		availabilityRepo: availabilityRepo,
		locationRepo:     locationRepo,
		bookingRepo:      bookingRepo,
		userRepo:         userRepo,
	}
}

// UpdateProfileRequest 更新老师资料请求
type UpdateProfileRequest struct {
	UserID          uint64   `json:"user_id"`
	Bio             string   `json:"bio"`
	Specialties     []string `json:"specialties"`
	TeachingYears   int      `json:"teaching_years"`
	HourlyRate      float64  `json:"hourly_rate"`
	TeachingStyle   string   `json:"teaching_style"`
	Certifications  []string `json:"certifications"`
	Languages       []string `json:"languages"`
}

// UpdateProfile 更新老师资料
func (s *TeacherService) UpdateProfile(ctx context.Context, req *UpdateProfileRequest) (*model.TeacherProfile, error) {
	profile := &model.TeacherProfile{
		UserID:         req.UserID,
		Bio:            req.Bio,
		Specialties:    req.Specialties,
		TeachingYears:  req.TeachingYears,
		HourlyRate:     req.HourlyRate,
		TeachingStyle:  req.TeachingStyle,
		Certifications: req.Certifications,
		Languages:      req.Languages,
	}

	if err := s.profileRepo.CreateOrUpdate(ctx, profile); err != nil {
		return nil, err
	}

	return s.profileRepo.GetByUserID(ctx, req.UserID)
}

// GetProfile 获取老师资料
func (s *TeacherService) GetProfile(ctx context.Context, userID uint64) (*model.TeacherProfile, error) {
	return s.profileRepo.GetByUserID(ctx, userID)
}

// SetAvailabilityRequest 设置可用时间请求
type SetAvailabilityRequest struct {
	TeacherID uint64     `json:"teacher_id"`
	Slots     []TimeSlot `json:"slots"`
}

// TimeSlot 时间段
type TimeSlot struct {
	Date      string `json:"date"`       // YYYY-MM-DD
	StartTime string `json:"start_time"` // HH:MM
	EndTime   string `json:"end_time"`   // HH:MM
}

// SetAvailability 设置老师可用时间
func (s *TeacherService) SetAvailability(ctx context.Context, req *SetAvailabilityRequest) error {
	// 先删除现有未来时间
	if err := s.availabilityRepo.DeleteByTeacherID(ctx, req.TeacherID); err != nil {
		return err
	}

	// 创建新的可用时间段
	var availabilities []*model.TeacherAvailability
	for _, slot := range req.Slots {
		availabilities = append(availabilities, &model.TeacherAvailability{
			TeacherID: req.TeacherID,
			Date:      slot.Date,
			StartTime: slot.StartTime,
			EndTime:   slot.EndTime,
			IsBooked:  false,
		})
	}

	if len(availabilities) > 0 {
		return s.availabilityRepo.CreateBatch(ctx, availabilities)
	}
	return nil
}

// GetAvailability 获取老师可用时间
func (s *TeacherService) GetAvailability(ctx context.Context, teacherID uint64, date string) ([]model.TeacherAvailability, error) {
	if date != "" {
		return s.availabilityRepo.GetByTeacherAndDate(ctx, teacherID, date)
	}
	return s.availabilityRepo.GetByTeacherID(ctx, teacherID)
}

// SetLocationsRequest 设置授课地点请求
type SetLocationsRequest struct {
	TeacherID uint64            `json:"teacher_id"`
	Locations []LocationRequest `json:"locations"`
}

// LocationRequest 地点请求
type LocationRequest struct {
	Type      model.LocationType `json:"type"`
	Name      string             `json:"name"`
	MetroLine string             `json:"metro_line,omitempty"`
	Address   string             `json:"address"`
	Longitude float64            `json:"longitude"`
	Latitude  float64            `json:"latitude"`
	RadiusKm  float64            `json:"radius_km"`
	IsDefault bool               `json:"is_default"`
}

// SetLocations 设置老师授课地点
func (s *TeacherService) SetLocations(ctx context.Context, req *SetLocationsRequest) error {
	// 先删除现有地点
	if err := s.locationRepo.DeleteByTeacherID(ctx, req.TeacherID); err != nil {
		return err
	}

	// 创建新的授课地点
	for _, loc := range req.Locations {
		location := &model.TeachingLocation{
			TeacherID: req.TeacherID,
			Type:      loc.Type,
			Name:      loc.Name,
			MetroLine: loc.MetroLine,
			Address:   loc.Address,
			Longitude: loc.Longitude,
			Latitude:  loc.Latitude,
			RadiusKm:  loc.RadiusKm,
			IsDefault: loc.IsDefault,
		}
		if err := s.locationRepo.Create(ctx, location); err != nil {
			return err
		}
	}
	return nil
}

// GetLocations 获取老师授课地点
func (s *TeacherService) GetLocations(ctx context.Context, teacherID uint64) ([]model.TeachingLocation, error) {
	return s.locationRepo.GetByTeacherID(ctx, teacherID)
}

// ListBookingRequests 获取老师的预约请求列表
func (s *TeacherService) ListBookingRequests(ctx context.Context, teacherID uint64, status string, page, pageSize int) ([]model.BookingRequest, int64, error) {
	return s.bookingRepo.ListByTeacher(ctx, teacherID, status, page, pageSize)
}

// AcceptBookingRequest 接受预约请求
func (s *TeacherService) AcceptBookingRequest(ctx context.Context, requestID uint64, responseNote string) error {
	return s.bookingRepo.UpdateStatus(ctx, requestID, model.BookingStatusAccepted, responseNote)
}

// RejectBookingRequest 拒绝预约请求
func (s *TeacherService) RejectBookingRequest(ctx context.Context, requestID uint64, responseNote string) error {
	return s.bookingRepo.UpdateStatus(ctx, requestID, model.BookingStatusRejected, responseNote)
}

// ===== Student Service =====

// StudentService 学生服务
type StudentService struct {
	bookingRepo      *repository.BookingRequestRepository
	teacherRepo      *repository.TeacherProfileRepository
	locationRepo     *repository.TeachingLocationRepository
	availabilityRepo *repository.TeacherAvailabilityRepository
	userRepo         *repository.UserRepository
}

func NewStudentService(
	bookingRepo *repository.BookingRequestRepository,
	teacherRepo *repository.TeacherProfileRepository,
	locationRepo *repository.TeachingLocationRepository,
	availabilityRepo *repository.TeacherAvailabilityRepository,
	userRepo *repository.UserRepository,
) *StudentService {
	return &StudentService{
		bookingRepo:      bookingRepo,
		teacherRepo:      teacherRepo,
		locationRepo:     locationRepo,
		availabilityRepo: availabilityRepo,
		userRepo:         userRepo,
	}
}

// NearbyTeacherRequest 搜索附近老师请求
type NearbyTeacherRequest struct {
	Latitude    float64 `json:"latitude"`    // 纬度
	Longitude   float64 `json:"longitude"`   // 经度
	RadiusKm    float64 `json:"radius_km"`   // 搜索半径（公里）
	Page        int     `json:"page"`
	PageSize    int     `json:"page_size"`
}

// NearbyTeacherResponse 附近老师响应
type NearbyTeacherResponse struct {
	TeacherID       uint64                  `json:"teacher_id"`
	Nickname        string                  `json:"nickname"`
	Avatar          string                  `json:"avatar"`
	Bio             string                  `json:"bio"`
	Specialties     []string                `json:"specialties"`
	TeachingYears   int                     `json:"teaching_years"`
	HourlyRate      float64                 `json:"hourly_rate"`
	Rating          float32                 `json:"rating"`
	TotalReviews    int                     `json:"total_reviews"`
	Distance        float64                 `json:"distance"`       // 距离（公里）
	Locations       []model.TeachingLocation `json:"locations"`     // 授课地点
}

// FindNearbyTeachers 查找附近的老师
func (s *StudentService) FindNearbyTeachers(ctx context.Context, req *NearbyTeacherRequest) ([]NearbyTeacherResponse, int64, error) {
	// 先获取附近的授课地点
	locations, err := s.locationRepo.FindNearbyLocations(ctx, req.Latitude, req.Longitude, req.RadiusKm)
	if err != nil {
		return nil, 0, err
	}

	// 按老师分组
	teacherLocations := make(map[uint64][]model.TeachingLocation)
	for _, loc := range locations {
		teacherLocations[loc.TeacherID] = append(teacherLocations[loc.TeacherID], loc)
	}

	// 获取老师详情
	var responses []NearbyTeacherResponse
	for teacherID, locs := range teacherLocations {
		profile, err := s.teacherRepo.GetByUserID(ctx, teacherID)
		if err != nil {
			continue
		}
		user, err := s.userRepo.GetByID(ctx, teacherID)
		if err != nil {
			continue
		}

		// 计算最小距离
		minDistance := math.MaxFloat64
		for _, loc := range locs {
			dist := calculateDistance(req.Latitude, req.Longitude, loc.Latitude, loc.Longitude)
			if dist < minDistance {
				minDistance = dist
			}
		}

		responses = append(responses, NearbyTeacherResponse{
			TeacherID:     teacherID,
			Nickname:      user.Nickname,
			Avatar:        user.Avatar,
			Bio:           profile.Bio,
			Specialties:   profile.Specialties,
			TeachingYears: profile.TeachingYears,
			HourlyRate:    profile.HourlyRate,
			Rating:        profile.Rating,
			TotalReviews:  profile.TotalReviews,
			Distance:      minDistance,
			Locations:     locs,
		})
	}

	return responses, int64(len(responses)), nil
}

// CreateBookingRequest 创建预约请求
type CreateBookingRequest struct {
	StudentID  uint64  `json:"student_id"`
	TeacherID  uint64  `json:"teacher_id"`
	LocationID uint64  `json:"location_id"`
	Date       string  `json:"date"`       // YYYY-MM-DD
	StartTime  string  `json:"start_time"` // HH:MM
	EndTime    string  `json:"end_time"`   // HH:MM
	Duration   int     `json:"duration"`   // 分钟
	Message    string  `json:"message"`
}

// CreateBookingRequest 学生创建预约请求
func (s *StudentService) CreateBookingRequest(ctx context.Context, req *CreateBookingRequest) (*model.BookingRequest, error) {
	// 生成请求单号
	requestNo := fmt.Sprintf("REQ%s%d", time.Now().Format("20060102"), time.Now().UnixNano()%1000000)

	// 获取老师资料确定价格
	profile, err := s.teacherRepo.GetByUserID(ctx, req.TeacherID)
	if err != nil {
		return nil, errors.New("老师不存在")
	}

	request := &model.BookingRequest{
		RequestNo:  requestNo,
		StudentID:  req.StudentID,
		TeacherID:  req.TeacherID,
		LocationID: req.LocationID,
		Date:       req.Date,
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
		Duration:   req.Duration,
		Status:     model.BookingStatusPending,
		Message:    req.Message,
		Price:      profile.HourlyRate * float64(req.Duration) / 60.0,
	}

	if err := s.bookingRepo.Create(ctx, request); err != nil {
		return nil, err
	}

	return s.bookingRepo.GetByID(ctx, request.ID)
}

// ListMyBookingRequests 获取学生的预约请求列表
func (s *StudentService) ListMyBookingRequests(ctx context.Context, studentID uint64, page, pageSize int) ([]model.BookingRequest, int64, error) {
	return s.bookingRepo.ListByStudent(ctx, studentID, page, pageSize)
}

// calculateDistance 计算两点间距离（Haversine公式）
func calculateDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371 // 地球半径（公里）

	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLng := (lng2 - lng1) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLng/2)*math.Sin(deltaLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

// ===== 以下保留原有代码用于兼容 =====

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

// CreateBookingRequestOld 创建预约请求（旧版兼容）
type CreateBookingRequestOld struct {
	UserID     uint64 `json:"user_id" binding:"required"`
	ScheduleID uint64 `json:"schedule_id" binding:"required"`
}

// CreateBooking 创建预约
func (s *BookingService) CreateBooking(ctx context.Context, req *CreateBookingRequestOld) (*model.Booking, error) {
	schedule, err := s.courseRepo.GetScheduleByID(ctx, req.ScheduleID)
	if err != nil {
		return nil, errors.New("课程不存在")
	}

	if schedule.BookedCount >= schedule.MaxStudents {
		return nil, errors.New("课程已满")
	}

	courseTime, _ := time.Parse("2006-01-02 15:04", schedule.CourseDate+" "+schedule.StartTime)
	if time.Now().After(courseTime) {
		return nil, errors.New("课程已开始")
	}

	bookingNo := fmt.Sprintf("BK%s%d", time.Now().Format("20060102"), time.Now().UnixNano()%1000000)

	booking := &model.Booking{
		BookingNo:   bookingNo,
		UserID:      req.UserID,
		ScheduleID:  req.ScheduleID,
		Status:      2,
		Price:       schedule.Price,
		CheckinCode: generateCheckinCode(),
		PaidAt:      &[]time.Time{time.Now()}[0],
	}

	if err := s.bookingRepo.Create(ctx, booking); err != nil {
		return nil, err
	}

	if err := s.courseRepo.UpdateBookedCount(ctx, req.ScheduleID, 1); err != nil {
		// 记录日志
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

	schedule, err := s.courseRepo.GetScheduleByID(ctx, booking.ScheduleID)
	if err != nil {
		return err
	}

	courseTime, _ := time.Parse("2006-01-02 15:04", schedule.CourseDate+" "+schedule.StartTime)
	if time.Until(courseTime) < 12*time.Hour {
		return errors.New("开课12小时内不可取消")
	}

	booking.Status = 0
	if err := s.bookingRepo.Update(ctx, booking); err != nil {
		return err
	}

	if err := s.courseRepo.UpdateBookedCount(ctx, booking.ScheduleID, -1); err != nil {
		// 记录日志
	}

	return nil
}

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
