package employee

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"gorm.io/gorm"
)

type SQLRepository struct {
	db *gorm.DB
}

func NewSQLRepository(db *gorm.DB) Repository {
	return &SQLRepository{db: db}
}

func (r SQLRepository) GetAllEmployees(ctx context.Context) (people []Employee, err error) {
	span, spanCtx := opentracing.StartSpanFromContext(ctx, "models.AllEmployees")
	defer span.Finish()

	err = r.db.WithContext(spanCtx).Find(&people).Error
	return
}

func (r SQLRepository) AddEmployee(ctx context.Context, newEmployee *Employee) (err error) {
	span, spanCtx := opentracing.StartSpanFromContext(ctx, "models.AddEmployee")
	defer span.Finish()

	err = r.db.WithContext(spanCtx).Create(newEmployee).Error
	return
}

func (r SQLRepository) GetEmployee(ctx context.Context, employeeID int64) (*Employee, error) {
	span, spanCtx := opentracing.StartSpanFromContext(ctx, "models.GetEmployee")
	defer span.Finish()

	employee := Employee{}
	err := r.db.WithContext(spanCtx).First(&employee, employeeID).Error
	return &employee, err
}

func (r SQLRepository) DeleteEmployee(ctx context.Context, employeeID int64) (err error) {
	span, spanCtx := opentracing.StartSpanFromContext(ctx, "models.DeleteEmployee")
	defer span.Finish()

	err = r.db.WithContext(spanCtx).Delete(&Employee{}, employeeID).Error
	return
}
