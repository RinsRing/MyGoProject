package main

import (
	"log"

	"MyGoProject/internal/handler"
	"MyGoProject/internal/middleware"
	"MyGoProject/internal/model"
	"MyGoProject/pkg/database"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	dsn := "root:123456@tcp(127.0.0.1:3306)/recruit_db?charset=utf8mb4&parseTime=True&loc=Local"
	database.InitDB(dsn)

	err := database.DB.AutoMigrate(&model.User{}, &model.Job{}, &model.Resume{})
	if err != nil {
		log.Fatalf("自动迁移失败: %v", err)
	}
	log.Println("数据表创建/更新成功")

	if err := database.InitRedis("localhost:6379", "", 0); err != nil {
		log.Fatalf("连接 Redis 失败: %v", err)
	}

	r.POST("/api/register", func(c *gin.Context) { handler.RegisterHandler(database.DB, c) })
	r.POST("/api/login", func(c *gin.Context) { handler.LogInHandler(database.DB, c) })

	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/user/info", func(c *gin.Context) {
			userID, _ := c.Get("userID")
			c.JSON(200, gin.H{"userID": userID})
		})

		protected.POST("/jobs", func(c *gin.Context) {
			handler.CreateJobHandler(database.DB, c, database.RDB)
		})

		protected.GET("/jobs", func(c *gin.Context) {
			handler.GetJobHandler(database.DB, c, database.RDB)
		})
		protected.POST("/resumes/upload", func(c *gin.Context) {
			handler.UploadResumeHandler(c, database.RDB)
		})
	}

	database.InitMQ()
	database.StartConsume()

	if err := r.Run(); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
