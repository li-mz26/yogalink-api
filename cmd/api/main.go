package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

// @title YogaLink API
// @version 1.0
// @description YogaLink 瑜伽约课平台 API 服务
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
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to YogaLink API",
			"version": "1.0.0",
		})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
		})
	})

	log.Println("Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
