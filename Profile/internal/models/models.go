// Package models
package models

import "github.com/shopspring/decimal"

type UserProfile struct {
	ID     uint   `gorm:"primaryKey"`
	UserID uint   `gorm:"not null"`
	Coins  []Coin `gorm:"foreignKey:UserProfileID"`
}

type Coin struct {
	ID       uint            `gorm:"primaryKey"`
	Symbol   string          `gorm:"not null"`
	Quantity decimal.Decimal `gorm:"type:decimal(20,8);not null"`

	UserProfileID uint
	UserProfile   UserProfile `gorm:"constraint:OnDelete:CASCADE;"`
}

type UserRequest struct {
	Name     string `json:"name"     binding:"required"`
	Password string `json:"password" binding:"required"`
}
