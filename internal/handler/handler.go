package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"MyGoProject/internal/auth"
	"MyGoProject/internal/model"
	"MyGoProject/pkg/database"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func RegisterHandler(db *gorm.DB, c *gin.Context) {
	var u model.RegisterRequest
	var existingUser model.User
	if err := c.ShouldBindJSON(&u); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := db.Where("phone_number = ?", u.PhoneNumber).First(&existingUser).Error; err == nil {
		c.JSON(400, gin.H{"error": "该手机号已注册"})
		return
	}
	hashPassword, pserr := auth.HashPassword(u.Password)
	if pserr != nil {
		c.JSON(400, gin.H{"error": "密码有问题"})
		return
	}
	newUser := model.User{
		Name:        u.Name,
		PhoneNumber: u.PhoneNumber,
		Age:         u.Age,
		Password:    hashPassword,
	}
	result := db.Create(&newUser)
	if result.Error != nil {
		c.JSON(500, gin.H{"error": "数据库写入失败"})
		return
	}
	c.JSON(200, gin.H{"message": "注册成功", "user": u.Name})
}

func LogInHandler(db *gorm.DB, c *gin.Context) {
	var currentUser model.LogRequest
	var existingUser model.User
	if err := c.ShouldBindJSON(&currentUser); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := db.Where("phone_number = ?", currentUser.PhoneNumber).First(&existingUser).Error; err != nil {
		c.JSON(400, gin.H{"error": "用户不存在"})
		return
	}
	if !auth.CheckPasswordHash(currentUser.Password, existingUser.Password) {
		c.JSON(400, gin.H{"error": "密码错误"})
		return
	}
	token, err := auth.GenerateToken(existingUser.ID)
	if err != nil {
		c.JSON(400, gin.H{"error": "token生成失败"})
		return
	}
	c.JSON(200, gin.H{"message": "登录成功", "token": token})
}

func CreateJobHandler(db *gorm.DB, c *gin.Context, rdb *redis.Client) {
	var cJob model.CreateJobRequest
	if err := c.ShouldBindJSON(&cJob); err != nil {
		c.JSON(400, gin.H{"error": "发布失败"})
		return
	}
	newJob := model.Job{
		Title:    cJob.Title,
		Describe: cJob.Describe,
	}
	result := db.Create(&newJob)
	if result.Error != nil {
		c.JSON(500, gin.H{"error": "岗位写入数据库失败"})
		return
	}
	rdb.Del(c.Request.Context(), "api:job_list")
	c.JSON(200, gin.H{"message": "岗位发布成功", "job": cJob.Title})
}

func GetJobHandler(db *gorm.DB, c *gin.Context, rdb *redis.Client) {
	var jobs []model.Job
	cacheKey := "api:job_list"
	ctx := c.Request.Context()
	val, err := rdb.Get(ctx, cacheKey).Result()
	if err == redis.Nil {
		result := db.Find(&jobs)
		if result.Error != nil {
			c.JSON(500, gin.H{"error": "岗位查找失败"})
			return
		}
		byteData, err := json.Marshal(jobs)
		if err != nil {
			c.JSON(500, gin.H{"error": "序列化失败"})
			return
		}
		randomMinutes := rand.Intn(5) + 1
		finalTime := time.Duration(10+randomMinutes) * time.Minute
		if err := rdb.Set(ctx, cacheKey, byteData, finalTime).Err(); err != nil {
			c.JSON(500, gin.H{"error": "写入redis失败"})
			return
		}
		c.JSON(200, gin.H{"data": jobs})
	} else if err != nil {
		c.JSON(500, gin.H{"error": "redis崩溃"})
		return
	} else {
		if err := json.Unmarshal([]byte(val), &jobs); err != nil {
			c.JSON(500, gin.H{"error": "反序列化失败"})
			return
		}
		c.JSON(200, gin.H{"data": jobs})
	}
}

func UploadResumeHandler(c *gin.Context, rdb *redis.Client) {
	userID, _ := c.Get("userID")
	limitKey := fmt.Sprintf("limit:resume:%d", userID)
	success, err := rdb.SetNX(c.Request.Context(), limitKey, "1", 1*time.Minute).Result()
	if err != nil {
		c.JSON(500, gin.H{"error": "系统繁忙"})
		return
	}
	if !success {
		c.JSON(429, gin.H{"error": "投递太频繁了"})
		return
	}
	var uploadResume model.ResumeMessage
	if err := c.ShouldBindJSON(&uploadResume); err != nil {
		c.JSON(400, gin.H{"error": "提交失败"})
		return
	}
	userid, _ := c.Get("userID")
	msg := model.ResumeMessage{
		UserID:  userid.(uint),
		JobID:   uploadResume.JobID,
		Content: uploadResume.Content,
	}
	body, err := json.Marshal(msg)
	if err != nil {
		c.JSON(500, gin.H{"error": "结构体转换JSON失败"})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = database.MQChannel.PublishWithContext(ctx,
		"",
		"resume_queue",
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		},
	)
	if err != nil {
		c.JSON(500, gin.H{"error": "投递失败"})
		return
	}

	c.JSON(200, gin.H{"message": "提交成功"})
}
