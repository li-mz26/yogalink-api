package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/yogalink/yogalink-api/internal/handler"
	"github.com/yogalink/yogalink-api/internal/repository"
	"github.com/yogalink/yogalink-api/internal/service"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// @title YogaLink API
// @version 2.0
// @description YogaLink 瑜伽老师-学生撮合平台 API 服务
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url https://yogalink.com
// @contact.email support@yogalink.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host api.yogalink.com
// @BasePath /v1
// @schemes https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description 请输入 JWT Token，格式: Bearer {token}

func main() {
	// 初始化数据库连接
	db, err := initDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 初始化仓库层
	userRepo := repository.NewUserRepository(db)
	profileRepo := repository.NewTeacherProfileRepository(db)
	availabilityRepo := repository.NewTeacherAvailabilityRepository(db)
	locationRepo := repository.NewTeachingLocationRepository(db)
	bookingRequestRepo := repository.NewBookingRequestRepository(db)
	courseRepo := repository.NewCourseRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	studioRepo := repository.NewStudioRepository(db)

	// 初始化服务层
	userService := service.NewUserService(userRepo)
	teacherService := service.NewTeacherService(profileRepo, availabilityRepo, locationRepo, bookingRequestRepo, userRepo)
	studentService := service.NewStudentService(bookingRequestRepo, profileRepo, locationRepo, availabilityRepo, userRepo)
	courseService := service.NewCourseService(courseRepo, bookingRepo)
	bookingService := service.NewBookingService(bookingRepo, courseRepo)
	studioService := service.NewStudioService(studioRepo)

	// 初始化处理器
	userHandler := handler.NewUserHandler(userService)
	teacherHandler := handler.NewTeacherHandler(teacherService)
	studentHandler := handler.NewStudentHandler(studentService)
	courseHandler := handler.NewCourseHandler(courseService)
	bookingHandler := handler.NewBookingHandler(bookingService)
	studioHandler := handler.NewStudioHandler(studioService)

	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to YogaLink API",
			"version": "2.0.0",
		})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
		})
	})

	// API 路由组
	v1 := r.Group("/v1")
	{
		// 公开路由
		v1.POST("/auth/register", userHandler.Register)
		v1.POST("/auth/login", userHandler.Login)

		// 需要认证的路由
		auth := v1.Group("/")
		auth.Use(handler.JWTAuth())
		{
			// 用户相关
			auth.GET("/user/profile", userHandler.GetProfile)

			// 老师端 API
			teacher := auth.Group("/teacher")
			{
				teacher.GET("/profile", teacherHandler.GetProfile)
				teacher.POST("/profile", teacherHandler.UpdateProfile)
				teacher.GET("/availability", teacherHandler.GetAvailability)
				teacher.POST("/availability", teacherHandler.SetAvailability)
				teacher.GET("/locations", teacherHandler.GetLocations)
				teacher.POST("/locations", teacherHandler.SetLocations)
				teacher.GET("/booking-requests", teacherHandler.ListBookingRequests)
				teacher.POST("/booking-requests/:id/accept", teacherHandler.AcceptBookingRequest)
				teacher.POST("/booking-requests/:id/reject", teacherHandler.RejectBookingRequest)
			}

			// 学生端 API
			student := auth.Group("/student")
			{
				student.POST("/booking-request", studentHandler.CreateBookingRequest)
				student.GET("/booking-requests", studentHandler.ListMyBookingRequests)
			}

			// 公开搜索 API
			v1.GET("/teachers/nearby", studentHandler.FindNearbyTeachers)

			// 兼容旧版 API（可选）
			auth.GET("/courses", courseHandler.ListSchedules)
			auth.GET("/courses/:id", courseHandler.GetScheduleDetail)
			auth.POST("/bookings", bookingHandler.CreateBooking)
			auth.GET("/bookings", bookingHandler.ListMyBookings)
			auth.GET("/bookings/:id", bookingHandler.GetBookingDetail)
			auth.POST("/bookings/:id/cancel", bookingHandler.CancelBooking)
			auth.GET("/studios", studioHandler.ListStudios)
			auth.GET("/studios/:id", studioHandler.GetStudioDetail)
		}
	}

	log.Println("Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

func initDB() (*gorm.DB, error) {
	dsn := "root:password@tcp(127.0.0.1:3306)/yogalink?charset=utf8mb4&parseTime=True&loc=Local"
	// 实际生产环境应从配置文件读取
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}
