package handlers

import (
	"net/http"

	"github.com/Wikia/go-example-service/cmd/server/models"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/jinzhu/gorm"
)

type Employee struct {
	Id   int
	Name string
	City string
}

func All(db *gorm.DB) func(ctx *gin.Context) {
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
