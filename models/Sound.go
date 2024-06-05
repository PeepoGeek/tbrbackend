package models

import "gorm.io/gorm"

type Sound struct {
	gorm.Model

	UUID        string `gorm:"primaryKey;type:varchar(36);not null;uniqueIndex" json:"uuid"`
	Name        string `gorm:"type:varchar(100);not null;uniqueIndex" json:"name"`
	Type        string `gorm:"type:varchar(50);not null;default:'background'" json:"type"` // 'background' or 'session'
	Tier        string `gorm:"type:varchar(50);not null;default:'free'" json:"tier"`       // 'free' or 'premium'
	Duration    int    `gorm:"not null" json:"duration"`                                   // duration in seconds
	Description string `gorm:"type:text" json:"description"`
	URL         string `gorm:"type:varchar(255);not null" json:"url"`
}
