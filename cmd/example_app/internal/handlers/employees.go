package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Wikia/go-example-service/cmd/example_app/internal/models"

	"github.com/jinzhu/gorm"

	"github.com/harnash/go-middlewares/logging"
)

type Employee struct {
	Id   int
	Name string
	City string
}

func All(db *gorm.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.FromRequest(r)
		logger.Info("Fetching list of all employees")

		people, err := models.AllEmployees(db)

		if err != nil {
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
