package httpapi

import (
	"go-workbench/internal/task"

	"github.com/gin-gonic/gin"
)

func NewRouter(service *task.Service, middlewares ...gin.HandlerFunc) *gin.Engine {
	router := gin.New()
	router.Use(middlewares...)
	router.GET("/v1/tasks", listTasks(service))
	return router
}
