package service

import "time"

// UserSession user sessions (_session)
type UserSession struct {
	ID           string    `json:"id,omitempty" gorm:"size:255;primaryKey"`
	UserID       string    `json:"user_id,omitempty" gorm:"size:255;uniqueIndex:idx_user_sessions_user_id_created_at,priority:1"`
	DeviceID     string    `json:"device_id,omitempty" gorm:"size:255"`
	CreatedAt    time.Time `json:"created_at" gorm:"uniqueIndex:idx_user_sessions_user_id_created_at,priority:2"`
	LastAccessed time.Time `json:"last_accessed"`
}
