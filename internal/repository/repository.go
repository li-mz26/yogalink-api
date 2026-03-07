package repository

import (
	"context"
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
