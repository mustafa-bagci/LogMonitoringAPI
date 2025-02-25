package models

import "gorm.io/gorm"

type Log struct {
	gorm.Model
	Level   string `json:"level"`
	Message string `json:"message"`
	Service string `json:"service"`
}
