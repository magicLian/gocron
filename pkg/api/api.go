package api

import "github.com/gin-gonic/gin"

func (hs *HTTPServer) apiRegister() {
	r := hs.ginEngine

	r.Use(gin.Recovery())
	gin.DisableConsoleColor()

	v2 := r.Group("/v1")

	nodeRouter := v2.Group("/nodes")
	{
		nodeRouter.GET("/", nil)
		nodeRouter.DELETE("/:nodeId", nil)
	}

	taskRouter := v2.Group("/tasks")
	{
		taskRouter.POST("/", nil)
		taskRouter.GET("/", nil)
		taskRouter.GET("/:taskId", nil)
		taskRouter.DELETE("/:taskId", nil)
	}

}
