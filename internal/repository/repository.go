package repository

import (
	"context"
	"math"

	"github.com/yogalink/yogalink-api/internal/model"
	"gorm.io/gorm"
)

// UserRepository 用户数据访问
type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create 创建用户
func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// GetByID 根据ID获取用户
func (r *UserRepository) GetByID(ctx context.Context, id uint64) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByPhone 根据手机号获取用户
func (r *UserRepository) GetByPhone(ctx context.Context, phone string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("phone = ?", phone).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update 更新用户信息
func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// TeacherProfileRepository 老师档案数据访问
type TeacherProfileRepository struct {
	db *gorm.DB
}

func NewTeacherProfileRepository(db *gorm.DB) *TeacherProfileRepository {
	return &TeacherProfileRepository{db: db}
}

// CreateOrUpdate 创建或更新老师档案
func (r *TeacherProfileRepository) CreateOrUpdate(ctx context.Context, profile *model.TeacherProfile) error {
	var existing model.TeacherProfile
	err := r.db.WithContext(ctx).Where("user_id = ?", profile.UserID).First(&existing).Error
	if err == nil {
		// 存在则更新
		profile.ID = existing.ID
		return r.db.WithContext(ctx).Save(profile).Error
	}
	// 不存在则创建
	return r.db.WithContext(ctx).Create(profile).Error
}

// GetByUserID 根据用户ID获取老师档案
func (r *TeacherProfileRepository) GetByUserID(ctx context.Context, userID uint64) (*model.TeacherProfile, error) {
	var profile model.TeacherProfile
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&profile).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

// ListNearbyTeachers 获取附近的老师列表
func (r *TeacherProfileRepository) ListNearbyTeachers(ctx context.Context, lat, lng float64, radiusKm float64, page, pageSize int) ([]model.TeacherProfile, int64, error) {
	var profiles []model.TeacherProfile
	var total int64

	// 先获取所有有地点信息的老师档案
	query := r.db.WithContext(ctx).Model(&model.TeacherProfile{}).
		Joins("JOIN teaching_locations ON teaching_locations.teacher_id = teacher_profiles.user_id").
		Where("teacher_profiles.is_verified = ?", true)

	query.Count(&total)
	err := query.Order("teacher_profiles.rating DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&profiles).Error

	return profiles, total, err
}

// TeacherAvailabilityRepository 老师可用时间数据访问
type TeacherAvailabilityRepository struct {
	db *gorm.DB
}

func NewTeacherAvailabilityRepository(db *gorm.DB) *TeacherAvailabilityRepository {
	return &TeacherAvailabilityRepository{db: db}
}

// Create 创建可用时间段
func (r *TeacherAvailabilityRepository) Create(ctx context.Context, availability *model.TeacherAvailability) error {
	return r.db.WithContext(ctx).Create(availability).Error
}

// CreateBatch 批量创建可用时间段
func (r *TeacherAvailabilityRepository) CreateBatch(ctx context.Context, availabilities []*model.TeacherAvailability) error {
	return r.db.WithContext(ctx).CreateInBatches(availabilities, len(availabilities)).Error
}

// GetByTeacherID 获取老师的所有可用时间
func (r *TeacherAvailabilityRepository) GetByTeacherID(ctx context.Context, teacherID uint64) ([]model.TeacherAvailability, error) {
	var availabilities []model.TeacherAvailability
	err := r.db.WithContext(ctx).
		Where("teacher_id = ? AND date >= CURDATE()", teacherID).
		Order("date ASC, start_time ASC").
		Find(&availabilities).Error
	return availabilities, err
}

// GetByTeacherAndDate 获取老师某天的可用时间
func (r *TeacherAvailabilityRepository) GetByTeacherAndDate(ctx context.Context, teacherID uint64, date string) ([]model.TeacherAvailability, error) {
	var availabilities []model.TeacherAvailability
	err := r.db.WithContext(ctx).
		Where("teacher_id = ? AND date = ? AND is_booked = ?", teacherID, date, false).
		Order("start_time ASC").
		Find(&availabilities).Error
	return availabilities, err
}

// DeleteByTeacherID 删除老师的所有可用时间（用于重置）
func (r *TeacherAvailabilityRepository) DeleteByTeacherID(ctx context.Context, teacherID uint64) error {
	return r.db.WithContext(ctx).
		Where("teacher_id = ? AND date >= CURDATE()", teacherID).
		Delete(&model.TeacherAvailability{}).Error
}

// TeachingLocationRepository 授课地点数据访问
type TeachingLocationRepository struct {
	db *gorm.DB
}

func NewTeachingLocationRepository(db *gorm.DB) *TeachingLocationRepository {
	return &TeachingLocationRepository{db: db}
}

// Create 创建授课地点
func (r *TeachingLocationRepository) Create(ctx context.Context, location *model.TeachingLocation) error {
	return r.db.WithContext(ctx).Create(location).Error
}

// GetByTeacherID 获取老师的所有授课地点
func (r *TeachingLocationRepository) GetByTeacherID(ctx context.Context, teacherID uint64) ([]model.TeachingLocation, error) {
	var locations []model.TeachingLocation
	err := r.db.WithContext(ctx).
		Where("teacher_id = ?", teacherID).
		Find(&locations).Error
	return locations, err
}

// GetByID 根据ID获取授课地点
func (r *TeachingLocationRepository) GetByID(ctx context.Context, id uint64) (*model.TeachingLocation, error) {
	var location model.TeachingLocation
	err := r.db.WithContext(ctx).First(&location, id).Error
	if err != nil {
		return nil, err
	}
	return &location, nil
}

// DeleteByTeacherID 删除老师的所有授课地点
func (r *TeachingLocationRepository) DeleteByTeacherID(ctx context.Context, teacherID uint64) error {
	return r.db.WithContext(ctx).
		Where("teacher_id = ?", teacherID).
		Delete(&model.TeachingLocation{}).Error
}

// FindNearbyLocations 查找附近的授课地点
func (r *TeachingLocationRepository) FindNearbyLocations(ctx context.Context, lat, lng float64, radiusKm float64) ([]model.TeachingLocation, error) {
	var locations []model.TeachingLocation
	
	// 使用Haversine公式计算距离
	// 简化版本：先找出大致范围内的地点
	latRange := radiusKm / 111.0 // 1度纬度约111公里
	lngRange := radiusKm / (111.0 * math.Cos(lat*math.Pi/180.0))
	
	err := r.db.WithContext(ctx).
		Where("latitude BETWEEN ? AND ?", lat-latRange, lat+latRange).
		Where("longitude BETWEEN ? AND ?", lng-lngRange, lng+lngRange).
		Find(&locations).Error
	
	return locations, err
}

// BookingRequestRepository 预约请求数据访问
type BookingRequestRepository struct {
	db *gorm.DB
}

func NewBookingRequestRepository(db *gorm.DB) *BookingRequestRepository {
	return &BookingRequestRepository{db: db}
}

// Create 创建预约请求
func (r *BookingRequestRepository) Create(ctx context.Context, request *model.BookingRequest) error {
	return r.db.WithContext(ctx).Create(request).Error
}

// GetByID 根据ID获取预约请求
func (r *BookingRequestRepository) GetByID(ctx context.Context, id uint64) (*model.BookingRequest, error) {
	var request model.BookingRequest
	err := r.db.WithContext(ctx).
		Preload("Student").
		Preload("Teacher").
		Preload("Location").
		First(&request, id).Error
	if err != nil {
		return nil, err
	}
	return &request, nil
}

// ListByStudent 获取学生的预约请求列表
func (r *BookingRequestRepository) ListByStudent(ctx context.Context, studentID uint64, page, pageSize int) ([]model.BookingRequest, int64, error) {
	var requests []model.BookingRequest
	var total int64

	query := r.db.WithContext(ctx).Model(&model.BookingRequest{}).
		Where("student_id = ?", studentID).
		Preload("Teacher").
		Preload("Location")

	query.Count(&total)
	err := query.Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&requests).Error

	return requests, total, err
}

// ListByTeacher 获取老师收到的预约请求列表
func (r *BookingRequestRepository) ListByTeacher(ctx context.Context, teacherID uint64, status string, page, pageSize int) ([]model.BookingRequest, int64, error) {
	var requests []model.BookingRequest
	var total int64

	query := r.db.WithContext(ctx).Model(&model.BookingRequest{}).
		Where("teacher_id = ?", teacherID).
		Preload("Student").
		Preload("Location")
	
	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)
	err := query.Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&requests).Error

	return requests, total, err
}

// UpdateStatus 更新预约请求状态
func (r *BookingRequestRepository) UpdateStatus(ctx context.Context, id uint64, status model.BookingStatus, responseNote string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if responseNote != "" {
		updates["response_note"] = responseNote
	}
	return r.db.WithContext(ctx).Model(&model.BookingRequest{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// ===== 以下保留原有代码用于兼容 =====

// CourseRepository 课程数据访问
type CourseRepository struct {
	db *gorm.DB
}

func NewCourseRepository(db *gorm.DB) *CourseRepository {
	return &CourseRepository{db: db}
}

// ListSchedules 获取课程排期列表
func (r *CourseRepository) ListSchedules(ctx context.Context, date string, page, pageSize int) ([]model.CourseSchedule, int64, error) {
	var schedules []model.CourseSchedule
	var total int64

	query := r.db.WithContext(ctx).Model(&model.CourseSchedule{}).
		Where("course_date = ? AND status = 1", date).
		Preload("Template").
		Preload("Studio").
		Preload("Instructor")

	query.Count(&total)
	err := query.Order("start_time ASC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&schedules).Error

	return schedules, total, err
}

// GetScheduleByID 获取课程排期详情
func (r *CourseRepository) GetScheduleByID(ctx context.Context, id uint64) (*model.CourseSchedule, error) {
	var schedule model.CourseSchedule
	err := r.db.WithContext(ctx).
		Preload("Template").
		Preload("Studio").
		Preload("Instructor").
		First(&schedule, id).Error
	if err != nil {
		return nil, err
	}
	return &schedule, nil
}

// UpdateBookedCount 更新已预约人数
func (r *CourseRepository) UpdateBookedCount(ctx context.Context, scheduleID uint64, delta int) error {
	return r.db.WithContext(ctx).Model(&model.CourseSchedule{}).
		Where("id = ?", scheduleID).
		UpdateColumn("booked_count", gorm.Expr("booked_count + ?", delta)).Error
}

// BookingRepository 预约数据访问
type BookingRepository struct {
	db *gorm.DB
}

func NewBookingRepository(db *gorm.DB) *BookingRepository {
	return &BookingRepository{db: db}
}

// Create 创建预约
func (r *BookingRepository) Create(ctx context.Context, booking *model.Booking) error {
	return r.db.WithContext(ctx).Create(booking).Error
}

// GetByID 获取预约详情
func (r *BookingRepository) GetByID(ctx context.Context, id uint64) (*model.Booking, error) {
	var booking model.Booking
	err := r.db.WithContext(ctx).
		Preload("Schedule.Template").
		Preload("Schedule.Studio").
		Preload("Schedule.Instructor").
		First(&booking, id).Error
	if err != nil {
		return nil, err
	}
	return &booking, nil
}

// ListByUser 获取用户预约列表
func (r *BookingRepository) ListByUser(ctx context.Context, userID uint64, page, pageSize int) ([]model.Booking, int64, error) {
	var bookings []model.Booking
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Booking{}).
		Where("user_id = ?", userID).
		Preload("Schedule.Template").
		Preload("Schedule.Studio")

	query.Count(&total)
	err := query.Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&bookings).Error

	return bookings, total, err
}

// Update 更新预约
func (r *BookingRepository) Update(ctx context.Context, booking *model.Booking) error {
	return r.db.WithContext(ctx).Save(booking).Error
}

// StudioRepository 场馆数据访问
type StudioRepository struct {
	db *gorm.DB
}

func NewStudioRepository(db *gorm.DB) *StudioRepository {
	return &StudioRepository{db: db}
}

// List 获取场馆列表
func (r *StudioRepository) List(ctx context.Context, page, pageSize int) ([]model.Studio, int64, error) {
	var studios []model.Studio
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Studio{}).Where("status = 1")
	query.Count(&total)
	err := query.Order("rating DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&studios).Error

	return studios, total, err
}

// GetByID 获取场馆详情
func (r *StudioRepository) GetByID(ctx context.Context, id uint64) (*model.Studio, error) {
	var studio model.Studio
	err := r.db.WithContext(ctx).First(&studio, id).Error
	if err != nil {
		return nil, err
	}
	return &studio, nil
}
