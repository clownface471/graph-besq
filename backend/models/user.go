package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username   string `gorm:"unique;not null" json:"username"`
	Password   string `gorm:"not null" json:"-"`
	Role       string `gorm:"not null" json:"role"`       // MANAGER, LEADER, OPERATOR
	Department string `json:"department"`                 // HAPUS tag gorm default
}