package models

import (
	"context"

	"github.com/opentracing/opentracing-go"
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
	span, spanCtx := opentracing.StartSpanFromContext(ctx, "models.AllEmployees")
	defer span.Finish()

	err = db.WithContext(spanCtx).Find(&people).Error
	return
}

func AddEmployee(ctx context.Context, db *gorm.DB, newEmployee *Employee) (err error) {
	span, spanCtx := opentracing.StartSpanFromContext(ctx, "models.AddEmployee")
	defer span.Finish()

	err = db.WithContext(spanCtx).Create(newEmployee).Error
	return
}

func GetEmployee(ctx context.Context, db *gorm.DB, employeeId string) (*Employee, error) {
	span, spanCtx := opentracing.StartSpanFromContext(ctx, "models.GetEmployee")
	defer span.Finish()

	employee := Employee{}
	err := db.WithContext(spanCtx).First(&employee, employeeId).Error
	return &employee, err
}

func DeleteEmployee(ctx context.Context, db *gorm.DB, employeeId string) (err error) {
	span, spanCtx := opentracing.StartSpanFromContext(ctx, "models.DeleteEmployee")
	defer span.Finish()

	err = db.WithContext(spanCtx).Delete(&Employee{}, employeeId).Error
	return
}