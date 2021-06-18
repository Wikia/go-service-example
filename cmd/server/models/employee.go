package models

import (
	"github.com/jinzhu/gorm"
)

type Employee struct {
	Id   int
	Name string
	City string
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

func GetEmployee(db *gorm.DB, employeeId string) (*Employee, error) {
	employee := Employee{}
	err := db.First(&employee, employeeId).Error
	return &employee, err
}

func DeleteEmployee(db *gorm.DB, employeeId string) (err error) {
	err = db.Delete(&Employee{}, employeeId).Error
	return
}