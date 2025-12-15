package model

import (
	"time"

	"gorm.io/gorm"
)

type Session struct {
	ID        string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	UserName  string         `gorm:"index;not null" json:"username"`
	Title     string         `gorm:"type:varchar(100)" json:"title"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	ModelType string         `gorm:"type:varchar(50);not null" json:"model_type"`
}

type SessionInfo struct {
	SessionID string `json:"sessionId"`
	Title     string `json:"name"`
	ModelType string `json:"modelType"`
}
