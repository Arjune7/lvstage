package models

import "time"

type Ad struct {
	ID        uint      `gorm:"primaryKey"`
	Title     string    `gorm:"size:255;not null" json:"title"`
	ImageURL  string    `gorm:"size:500;not null" json:"image_url"`
	TargetURL string    `gorm:"size:500;not null" json:"target_url"`
	Status    string    `gorm:"size:50;default:'active';index" json:"status"` // active, paused, archived
	CreatedAt time.Time `gorm:"autoCreateTime;index" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
