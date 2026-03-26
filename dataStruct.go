package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB // 数据库实例

// 密码加密
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// 密码验证
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

type User struct {
	gorm.Model
	Name        string `gorm:"not null"`
	Age         int    `gorm:"not null"`
	Password    string `gorm:"not null"`
	PhoneNumber string `gorm:"unique;not null;index"`
}

type Job struct {
	gorm.Model
	Title    string `gorm:"not null"`
	Describe string `gorm:"type:text"`
}

type Resume struct {
	gorm.Model
	UserID  uint
	JobID   uint
	Content string
	Status  int `gorm:"type:tinyint;default:1"` //1待处理，2评估中，3已完成
}

// 接受数据的结构体
type RegisterRequest struct {
	Name        string `json:"name" binding:"required"`
	Age         int    `json:"age"`
	Password    string `json:"password" binding:"required"`
	PhoneNumber string `json:"phone_number" binding:"required"`
}

// 注册接口
func RegisterHandler(db *gorm.DB, c *gin.Context) {
	var u RegisterRequest
	var existingUser User
	if err := c.ShouldBindJSON(&u); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	//查重
	if err := db.Where("phone_number = ?", u.PhoneNumber).First(&existingUser).Error; err == nil {
		c.JSON(400, gin.H{"error": "该手机号已注册"})
		return
	}
	hashPassword, pserr := HashPassword(u.Password)
	if pserr != nil {
		c.JSON(400, gin.H{"error": "密码有问题"})
		return
	}
	newUser := User{
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

// Token
var jwtSecret = []byte("my_secret_key")

func GenerateToken(userID uint) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}

	// 生成 Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// 接受结构体
type LogRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
	Password    string `json:"password" binding:"required"`
}

// 登录接口
func LogInHandler(db *gorm.DB, c *gin.Context) {
	var currentUser LogRequest
	var existingUser User
	//获取数据
	if err := c.ShouldBindJSON(&currentUser); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	//查找
	if err := db.Where("phone_number = ?", currentUser.PhoneNumber).First(&existingUser).Error; err != nil {
		c.JSON(400, gin.H{"error": "用户不存在"})
		return
	}
	//验证
	if !CheckPasswordHash(currentUser.Password, existingUser.Password) {
		c.JSON(400, gin.H{"error": "密码错误"})
		return
	}
	//生成JWT Token
	token, err := GenerateToken(existingUser.ID)
	if err != nil {
		c.JSON(400, gin.H{"error": "token生成失败"})
		return
	}
	c.JSON(200, gin.H{"message": "登录成功", "token": token})
}

// 验证器
func ParseToken(tokenString string) (uint, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return 0, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := uint(claims["user_id"].(float64))
		return userID, nil
	}
	return 0, fmt.Errorf("invalid token")
}

// 中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"error": "需要登陆才能访问"})
			c.Abort()
			return
		}
		userID, err := ParseToken(authHeader)
		if err != nil {
			c.JSON(401, gin.H{"error": "无效token"})
			c.Abort()
			return
		}
		c.Set("userID", userID)
		c.Next()
	}
}

// 岗位接收器
type CreateJobRequest struct {
	Title    string `gorm:"not nil" binding:"required"`
	Describe string `gorm:"type:text"`
}

// 发布岗位
func CreateJobHandler(db *gorm.DB, c *gin.Context) {
	var cJob CreateJobRequest
	if err := c.ShouldBindJSON(&cJob); err != nil {
		c.JSON(400, gin.H{"error": "发布失败"})
		return
	}
	newJob := Job{
		Title:    cJob.Title,
		Describe: cJob.Describe,
	}
	result := db.Create(&newJob)
	if result.Error != nil {
		c.JSON(500, gin.H{"error": "岗位写入数据库失败"})
		return
	}
	c.JSON(200, gin.H{"message": "岗位发布成功", "job": cJob.Title})
}

// 查询岗位
func GetJobHandler(db *gorm.DB, c *gin.Context) {
	var jobs []Job
	result := db.Find(&jobs)
	if result.Error != nil {
		c.JSON(500, gin.H{"error": "岗位查找失败"})
		return
	}
	c.JSON(200, gin.H{"data": jobs})
}

// 投递结构体
type SubmitResumeRequest struct {
	JobID   uint   `json:"job_id" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// 投递函数
func UploadResumeHandler(db *gorm.DB, c *gin.Context) {
	var uploadResume SubmitResumeRequest
	if err := c.ShouldBindJSON(&uploadResume); err != nil {
		c.JSON(400, gin.H{"error": "提交失败"})
		return
	}
	userid, _ := c.Get("userID")
	newResume := Resume{
		UserID:  userid.(uint),
		JobID:   uploadResume.JobID,
		Content: uploadResume.Content,
	}
	result := db.Create(&newResume)
	if result.Error != nil {
		c.JSON(500, gin.H{"error": "简历进入数据库失败"})
		return
	}
	c.JSON(200, gin.H{"message": "提交成功"})
}

func main() {
	r := gin.Default()

	dsn := "root:123456@tcp(127.0.0.1:3306)/recruit_db?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	// 把连接好的 db 赋值给全局变量
	DB = db

	fmt.Println("数据库连接成功")

	// 2. 自动迁移
	err = DB.AutoMigrate(&User{}, &Job{}, &Resume{})
	if err != nil {
		log.Fatalf("自动迁移失败: %v", err)
	}
	fmt.Println("数据表创建/更新成功")

	r.POST("/api/register", func(c *gin.Context) { RegisterHandler(db, c) })
	r.POST("/api/login", func(c *gin.Context) { LogInHandler(db, c) })
	protected := r.Group("/api")
	protected.Use(AuthMiddleware())
	{
		protected.GET("/user/info", func(c *gin.Context) {
			// 在这里获取中间件塞进去的 userID
			userID, _ := c.Get("userID")
			c.JSON(200, gin.H{"userID": userID})
		})

		protected.POST("/jobs", func(c *gin.Context) {
			CreateJobHandler(db, c)
		})

		protected.GET("/jobs", func(c *gin.Context) {
			GetJobHandler(db, c)
		})
		protected.POST("/resumes/upload", func(c *gin.Context) {
			UploadResumeHandler(db, c)
		})
	}

	r.Run()
}
