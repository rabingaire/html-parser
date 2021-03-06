package api

import (
	"github.com/gin-gonic/gin"
	"github.com/rabingaire/html-parser/middleware"
)

// Setup routes and gin engine
func Setup() *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORS())
	r.GET("/api/v1/info", GetPageInfo)
	return r
}
