package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/jinzhu/gorm"

	logmiddleware "github.com/harnash/go-middlewares/logger"
)

type Employee struct {
	Id   int
	Name string
	City string
}

func InitData(db *gorm.DB) {
	db.AutoMigrate(&Employee{})
	db.Create(&Employee{Name: "Przemek", City: "Olsztyn"})
	db.Create(&Employee{Name: "Łukasz", City: "Poznań"})
}

func All(db *gorm.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logmiddleware.FromRequest(r)
		logger.Info("Fetching single employee")

		var people []Employee
		if err := db.Find(&people).Error; err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			b, err := json.Marshal(people)
			if err != nil {
				panic(err) // no, not really
			}
			w.WriteHeader(http.StatusOK)
			_, err = w.Write(b)
		}

	}
}
