package handlers

import (
	"net/http"

	"github.com/Wikia/go-example-service/cmd/server/models"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/jinzhu/gorm"
)

func AllEmployees(db *gorm.DB) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		logger := zap.S()
		logger.Info("Fetching list of all employees")

		people, err := models.AllEmployees(db)

		if err != nil {
			ctx.AbortWithStatus(http.StatusInternalServerError)
		} else {
			ctx.JSON(http.StatusOK, people)
		}
	}
}

func CreateEmployee(db *gorm.DB) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		logger := zap.S()
		e := &models.Employee{}
		if err := ctx.Bind(e); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		logger.With("employee", e).Info("creating new employee")
		if err := models.AddEmployee(db, e); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.Status(http.StatusAccepted)
	}
}

func GetEmployee(db *gorm.DB) func (ctx *gin.Context) {
	return func(ctx *gin.Context) {
		logger := zap.S()
		employeeId := ctx.Param("id")
		logger.With("id", employeeId).Info("looking up employee")
		e, err := models.GetEmployee(db, employeeId)
		if err == gorm.ErrRecordNotFound {
			ctx.AbortWithStatus(http.StatusNotFound)
			return
		} else if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, e)
	}
}

func DeleteEmployee(db *gorm.DB) func (ctx *gin.Context) {
	return func(ctx *gin.Context) {
		logger := zap.S()
		employeeId := ctx.Param("id")
		logger.With("id", employeeId).Info("deleteing employee")
		err := models.DeleteEmployee(db, employeeId)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx.Status(http.StatusAccepted)
	}
}
