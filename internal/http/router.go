package http

import "github.com/gin-gonic/gin"

type Deps struct {
	AuthMiddleWare gin.HandlerFunc
	MeHandler      gin.HandlerFunc
}

func NewRouter(d Deps) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	v1 := r.Group("/api/v1")
	v1.Use(d.AuthMiddleWare)

	return r
}
