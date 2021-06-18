package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Wikia/go-example-service/cmd/example_app/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"

	"github.com/harnash/go-middlewares/logging"
)

type Employee struct {
	Id   int
	Name string
	City string
}

//--
// Error response payloads & renderers
//--

// ErrResponse renderer type for handling all sorts of errors.
//
// In the best case scenario, the excellent github.com/pkg/errors package
// helps reveal information on the error, setting it on Err, and in the Render()
// method, using it to set the application-specific error code in AppCode.
type ErrResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText string `json:"status"`          // user-level status message
	AppCode    int64  `json:"code,omitempty"`  // application-specific error code
	ErrorText  string `json:"error,omitempty"` // application-level error message, for debugging
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 400,
		StatusText:     "Invalid request.",
		ErrorText:      err.Error(),
	}
}

var ErrNotFound = &ErrResponse{HTTPStatusCode: 404, StatusText: "Resource not found."}

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

func CreateEmployee(db *gorm.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := zap.S()
		ctx := r.Context()
		e, ok := ctx.Value("employee").(*models.Employee)
		if !ok {
			render.Status(r, http.StatusBadRequest)
			return
		}
		logger.With("employee", e).Info("creating new employee")
		if err := models.AddEmployee(db, e); err != nil {
			render.JSON(w, r, ErrInvalidRequest(err))
			return
		}
		render.Status(r, http.StatusAccepted)
	}
}

func GetEmployee(db *gorm.DB) func (w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := zap.S()
		employeeId := chi.URLParam(r, "id")
		logger.With("id", employeeId).Info("looking up employee")
		e, err := models.GetEmployee(db, employeeId)
		if err != nil {
			render.JSON(w, r, ErrInvalidRequest(err))
			return
		}
		if e == nil {
			render.JSON(w, r, ErrNotFound)
			return
		}

		render.JSON(w, r, e)
	}
}

func DeleteEmployee(db *gorm.DB) func (w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := zap.S()
		employeeId := chi.URLParam(r,"id")
		logger.With("id", employeeId).Info("deleting employee")
		err := models.DeleteEmployee(db, employeeId)
		if err != nil {
			render.JSON(w, r, ErrInvalidRequest(err))
			return
		}

		render.Status(r, http.StatusAccepted)
	}
}
