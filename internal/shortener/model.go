package shortener

import "time"

type URL struct {
	ID          uint       `gorm:"primary_key" json:"id"`
	ShortCode   string     `gorm:"type:char(5);uniqueIndex;not null" json:"short_code"`
	OriginalURL string     `gorm:"type:text;not null" json:"original_url"`
	CreatedAt   time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
	ExpiresAt   *time.Time `gorm:"default:null" json:"expires_at,omitempty"`
	ClickCount  int64      `gorm:"not null;default:0" json:"click_count"`
}
