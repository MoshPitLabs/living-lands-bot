package models

import "time"

type WelcomeTemplate struct {
	ID        uint   `gorm:"primaryKey"`
	Message   string `gorm:"not null"`
	Weight    int    `gorm:"default:1"`
	Active    bool   `gorm:"default:true"`
	CreatedAt time.Time
}
