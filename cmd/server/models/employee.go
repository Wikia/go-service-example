package models

import (
	"github.com/jinzhu/gorm"
)

type Employee struct {
	Id   int `json:"id"`
	Name string `json:"name"`
	City string `json:"city"`
}

func InitData(db *gorm.DB) {
	db.AutoMigrate(&Employee{})
	db.Create(&Employee{Id: 1, Name: "Przemek", City: "Olsztyn"})
	db.Create(&Employee{Id: 2, Name: "Łukasz", City: "Poznań"})
}

func AllEmployees(db *gorm.DB) (people []Employee, err error) {
	err = db.Find(&people).Error
	return
}

func AddEmployee(db *gorm.DB, newEmployee *Employee) (err error) {
	err = db.Create(newEmployee).Error
	return
}