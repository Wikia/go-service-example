package database

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"context"

	"gorm.io/gorm"
)

type EmployeeDBModel struct {
	ID   int    `gorm:"column:id;primaryKey"`
	Name string `gorm:"column:name"`
	City string `gorm:"column:city"`
}

//counterfeiter:generate . Repository
type Repository interface {
	GetAllEmployees(ctx context.Context) ([]EmployeeDBModel, error)
	AddEmployee(ctx context.Context, newEmployee *EmployeeDBModel) error
	GetEmployee(ctx context.Context, employeeID int64) (*EmployeeDBModel, error)
	DeleteEmployee(ctx context.Context, employeeID int64) error
}

func InitData(db *gorm.DB) (err error) {
	err = db.AutoMigrate(&EmployeeDBModel{})
	if err != nil {
		return
	}

	db.Create(&EmployeeDBModel{Name: "Przemek", City: "Olsztyn"})
	db.Create(&EmployeeDBModel{Name: "Łukasz", City: "Poznań"})

	return
}
