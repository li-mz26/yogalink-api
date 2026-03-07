package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yogalink/yogalink-api/internal/handler"
	"github.com/yogalink/yogalink-api/internal/model"
	"github.com/yogalink/yogalink-api/internal/repository"
	"github.com/yogalink/yogalink-api/internal/service"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// 使用 SQLite 作为 MVP 数据库（方便快速启动）
	db, err := gorm.Open(sqlite.Open("yogalink.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect database:", err)
	}

	// 自动迁移
	db.AutoMigrate(
		&model.User{},
		&model.Studio{},
		&model.Instructor{},
		&model.CourseTemplate{},
		&model.CourseSchedule{},
		&model.Booking{},
	)

	// 初始化仓库
	userRepo := repository.NewUserRepository(db)
	courseRepo := repository.NewCourseRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	studioRepo := repository.NewStudioRepository(db)

	// 初始化服务
	userService := service.NewUserService(userRepo)
	courseService := service.NewCourseService(courseRepo, bookingRepo)
	bookingService := service.NewBookingService(bookingRepo, courseRepo)
	studioService := service.NewStudioService(studioRepo)

	// 初始化处理器
	userHandler := handler.NewUserHandler(userService)
	courseHandler := handler.NewCourseHandler(courseService)
	bookingHandler := handler.NewBookingHandler(bookingService)
	studioHandler := handler.NewStudioHandler(studioService)

	// 插入测试数据
	seedData(db)

	// 设置路由
	r := gin.Default()

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
		})
	})

	// API 路由组
	api := r.Group("/v1")
	{
		// 公开路由
		api.POST("/auth/register", userHandler.Register)
		api.POST("/auth/login", userHandler.Login)

		// 课程相关（公开）
		api.GET("/courses", courseHandler.ListSchedules)
		api.GET("/courses/:id", courseHandler.GetScheduleDetail)

		// 场馆相关（公开）
		api.GET("/studios", studioHandler.ListStudios)
		api.GET("/studios/:id", studioHandler.GetStudioDetail)

		// 需要登录的路由
		authorized := api.Group("", handler.JWTAuth())
		{
			// 用户
			authorized.GET("/users/me", userHandler.GetProfile)

			// 预约
			authorized.POST("/bookings", bookingHandler.CreateBooking)
			authorized.GET("/bookings", bookingHandler.ListMyBookings)
			authorized.GET("/bookings/:id", bookingHandler.GetBookingDetail)
			authorized.POST("/bookings/:id/cancel", bookingHandler.CancelBooking)
		}
	}

	log.Println("🚀 YogaLink API Server starting on :8080")
	log.Println("📚 API Documentation: http://localhost:8080/v1/courses")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

// seedData 插入测试数据
func seedData(db *gorm.DB) {
	// 检查是否已有数据
	var count int64
	db.Model(&model.Studio{}).Count(&count)
	if count > 0 {
		return
	}

	log.Println("🌱 Seeding test data...")

	// 创建场馆
	studios := []model.Studio{
		{
			Name:      "静心瑜伽馆",
			Address:   "上海市静安区南京西路1266号",
			Phone:     "021-12345678",
			Rating:    4.9,
			Status:    1,
			CreatedAt: time.Now(),
		},
		{
			Name:      "悦动瑜伽空间",
			Address:   "上海市徐汇区淮海中路999号",
			Phone:     "021-87654321",
			Rating:    4.7,
			Status:    1,
			CreatedAt: time.Now(),
		},
	}
	db.Create(&studios)

	// 创建教练
	instructors := []model.Instructor{
		{
			StudioID:      studios[0].ID,
			Name:          "李老师",
			Bio:           "10年瑜伽教学经验，擅长流瑜伽和阴瑜伽",
			Specialties:   []string{"流瑜伽", "阴瑜伽"},
			TeachingYears: 10,
			Rating:        4.9,
		},
		{
			StudioID:      studios[0].ID,
			Name:          "王老师",
			Bio:           "RYT500认证教练，专注哈他瑜伽",
			Specialties:   []string{"哈他瑜伽"},
			TeachingYears: 5,
			Rating:        4.8,
		},
	}
	db.Create(&instructors)

	// 创建课程模板
	templates := []model.CourseTemplate{
		{
			StudioID:     studios[0].ID,
			InstructorID: instructors[0].ID,
			Name:         "晨间流瑜伽",
			Style:        "流瑜伽",
			Level:        2,
			Duration:     60,
			MaxStudents:  15,
			Description:  "通过流畅的体式串联，唤醒身体能量",
			Price:        69,
		},
		{
			StudioID:     studios[0].ID,
			InstructorID: instructors[1].ID,
			Name:         "阴瑜伽深度放松",
			Style:        "阴瑜伽",
			Level:        1,
			Duration:     75,
			MaxStudents:  12,
			Description:  "长时间保持体式，深层放松筋膜",
			Price:        79,
		},
		{
			StudioID:     studios[1].ID,
			InstructorID: instructors[0].ID,
			Name:         "哈他基础",
			Style:        "哈他瑜伽",
			Level:        1,
			Duration:     60,
			MaxStudents:  20,
			Description:  "适合初学者的瑜伽基础课程",
			Price:        59,
		},
	}
	db.Create(&templates)

	// 创建课程排期（未来7天）
	baseDate := time.Now()
	for i := 0; i < 7; i++ {
		courseDate := baseDate.AddDate(0, 0, i).Format("2006-01-02")

		schedules := []model.CourseSchedule{
			{
				TemplateID:   templates[0].ID,
				StudioID:     studios[0].ID,
				InstructorID: instructors[0].ID,
				CourseDate:   courseDate,
				StartTime:    "09:00",
				EndTime:      "10:00",
				MaxStudents:  15,
				Price:        69,
				Status:       1,
			},
			{
				TemplateID:   templates[0].ID,
				StudioID:     studios[0].ID,
				InstructorID: instructors[0].ID,
				CourseDate:   courseDate,
				StartTime:    "18:30",
				EndTime:      "19:30",
				MaxStudents:  15,
				Price:        79,
				Status:       1,
			},
			{
				TemplateID:   templates[1].ID,
				StudioID:     studios[0].ID,
				InstructorID: instructors[1].ID,
				CourseDate:   courseDate,
				StartTime:    "20:00",
				EndTime:      "21:15",
				MaxStudents:  12,
				Price:        79,
				Status:       1,
			},
		}
		db.Create(&schedules)
	}

	log.Println("✅ Test data seeded successfully!")
}
