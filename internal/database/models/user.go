package models

import "time"

// User represents a Discord user account linked to a Hytale account.
type User struct {
	ID               uint   `gorm:"primaryKey"`
	DiscordID        string `gorm:"uniqueIndex;not null;type:varchar(20)"` // Discord IDs as strings (native discordgo format)
	DiscordUsername  string `gorm:"not null"`
	HytaleUsername   string `gorm:"index"`
	HytaleUUID       string `gorm:"index"`
	VerificationCode string `gorm:"index"`
	VerifiedAt       *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
