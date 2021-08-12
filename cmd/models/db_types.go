package models

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"context"

	"gorm.io/gorm"
)

type EmployeeDbModel struct {
	ID   int    `gorm:"column:id;primaryKey"`
	Name string `gorm:"column:name"`
	City string `gorm:"column:city"`
}

//counterfeiter:generate . Repository
type Repository interface {
	GetAllEmployees(ctx context.Context) ([]EmployeeDbModel, error)
	AddEmployee(ctx context.Context, newEmployee *EmployeeDbModel) error
	GetEmployee(ctx context.Context, employeeID int64) (*EmployeeDbModel, error)
	DeleteEmployee(ctx context.Context, employeeID int64) error
}

func InitData(db *gorm.DB) (err error) {
	err = db.AutoMigrate(&EmployeeDbModel{})
	if err != nil {
		return
	}

	db.Create(&EmployeeDbModel{Name: "Przemek", City: "Olsztyn"})
	db.Create(&EmployeeDbModel{Name: "Łukasz", City: "Poznań"})

	return
}
