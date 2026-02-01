package models

import "time"

type ChannelRoute struct {
	ID          uint   `gorm:"primaryKey"`
	Keyword     string `gorm:"uniqueIndex;not null"`
	ChannelID   string `gorm:"not null"`
	Description string
	Emoji       string
	CreatedAt   time.Time
}
