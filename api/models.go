package api

type CreateEmployeeRequest struct {
	Name string `json:"name" validate:"required,gt=3,lt=50,alphanumunicode|printascii"`
	City string `json:"city" validate:"required,gt=4,lt=30,alphanumunicode|printascii"`
}

type EmployeeResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	City string `json:"city"`
}
