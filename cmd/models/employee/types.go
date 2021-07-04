package employee

import (
	"context"

	"gorm.io/gorm"
)

type Employee struct {
	ID   int
	Name string `validate:"required,gt=3"`
	City string `validate:"required,gt=4"`
}

type Repository interface {
	GetAllEmployees(ctx context.Context) ([]Employee, error)
	AddEmployee(ctx context.Context, newEmployee *Employee) error
	GetEmployee(ctx context.Context, employeeID int64) (*Employee, error)
	DeleteEmployee(ctx context.Context, employeeID int64) error
}

func InitData(db *gorm.DB) (err error) {
	err = db.AutoMigrate(&Employee{})
	if err != nil {
		return
	}
	db.Create(&Employee{ID: 1, Name: "Przemek", City: "Olsztyn"})
	db.Create(&Employee{ID: 2, Name: "Łukasz", City: "Poznań"})

	return
}
