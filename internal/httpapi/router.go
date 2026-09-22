package httpapi

import (
	"go-workbench/internal/task"
	"net/http"

	"github.com/gin-gonic/gin"
)

func recoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				c.AbortWithStatusJSON(
					http.StatusInternalServerError,
					errorResponse{
						Code:    "INTERNAL_ERROR",
						Message: "internal server error",
					},
				)
			}
		}()

		c.Next()
	}
}

func NewRouter(service *task.Service, middlewares ...gin.HandlerFunc) *gin.Engine {
	router := gin.New()
	router.Use(recoveryMiddleware())
	router.Use(middlewares...)
	router.GET("/v1/tasks", listTasks(service))
	return router
}
