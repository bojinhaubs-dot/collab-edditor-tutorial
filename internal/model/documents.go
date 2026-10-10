package model

type Document struct {
	ID      uint64 `gorm:"primaryKey"`
	Title   string `gorm:"type:varchar(255);not null"`
	Content string `gorm:"type:longtext;not null"`
}
