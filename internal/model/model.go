package model

import "gorm.io/gorm"

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
	Status  int `gorm:"type:tinyint;default:1"` // 1待处理，2评估中，3已完成
}

// RegisterRequest 注册请求体
type RegisterRequest struct {
	Name        string `json:"name" binding:"required"`
	Age         int    `json:"age"`
	Password    string `json:"password" binding:"required"`
	PhoneNumber string `json:"phone_number" binding:"required"`
}

// LogRequest 登录请求体
type LogRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
	Password    string `json:"password" binding:"required"`
}

// CreateJobRequest 发布岗位请求体
type CreateJobRequest struct {
	Title    string `json:"title" binding:"required"`
	Describe string `json:"describe"`
}

// ResumeMessage 投递简历（队列消息）
type ResumeMessage struct {
	UserID  uint   `json:"user_id"`
	JobID   uint   `json:"job_id" binding:"required"`
	Content string `json:"content" binding:"required"`
}
