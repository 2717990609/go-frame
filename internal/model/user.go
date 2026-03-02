// Package model PO/DO 定义，Repository 层 DB 映射（规范 5.2）
package model

import (
	"go-backend-framework/internal/dto"
	"time"
)

// User 用户 PO，对应 sd_user 表
type User struct {
	ID             int64     `gorm:"column:id;primaryKey"`
	Email          string    `gorm:"column:email"`
	Phone          string    `gorm:"column:phone"`
	PasswordHash   string    `gorm:"column:password_hash"`
	Nickname       string    `gorm:"column:nickname"`
	Avatar         string    `gorm:"column:avatar"`
	Spark          float64   `gorm:"column:spark"`           // 星火余额
	Status         int       `gorm:"column:status"`
	LastLoginAt    *time.Time `gorm:"column:last_login_at"`
	LastLoginIP    string    `gorm:"column:last_login_ip"`
	DeviceID       string    `gorm:"column:device_id"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
}

// TableName 表名
func (User) TableName() string {
	return "sd_user"
}

// ToVO 转换为 VO，脱敏（规范 5.2）
func (u *User) ToVO() dto.UserVO {
	var createdAt int64
	if !u.CreatedAt.IsZero() {
		createdAt = u.CreatedAt.Unix()
	}
	return dto.UserVO{
		ID:        u.ID,
		Nickname:  u.Nickname,
		Avatar:    u.Avatar,
		Spark:     u.Spark,
		Status:    u.Status,
		CreatedAt: createdAt,
	}
}
