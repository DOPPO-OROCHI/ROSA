package player

import "time"

type AuthSession struct {
	ID        uint      `gorm:"primaryKey"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
	TokenHash string    `gorm:"size:64;uniqueIndex;not null"`
	UserID    uint      `gorm:"index;not null"`
	TGID      int64     `gorm:"index;not null;default:0"`
	ExpiresAt time.Time `gorm:"index;not null"`
}
