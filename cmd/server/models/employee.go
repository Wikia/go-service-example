package models

import (
	"context"

	"gorm.io/gorm"
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

func AllEmployees(ctx context.Context, db *gorm.DB) (people []Employee, err error) {
	err = db.WithContext(ctx).Find(&people).Error
	return
}

func AddEmployee(ctx context.Context, db *gorm.DB, newEmployee *Employee) (err error) {
	err = db.WithContext(ctx).Create(newEmployee).Error
	return
}

func GetEmployee(ctx context.Context, db *gorm.DB, employeeId string) (*Employee, error) {
	employee := Employee{}
	err := db.WithContext(ctx).First(&employee, employeeId).Error
	return &employee, err
}

func DeleteEmployee(ctx context.Context, db *gorm.DB, employeeId string) (err error) {
	err = db.WithContext(ctx).Delete(&Employee{}, employeeId).Error
	return
}