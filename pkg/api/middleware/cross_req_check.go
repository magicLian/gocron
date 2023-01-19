package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func CrossRequestCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		passed := false

		c.Writer.Header().Add("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Add("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization,authorization")
		c.Writer.Header().Add("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Add("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if c.Request.Method == "OPTIONS" && passed {
			c.JSON(http.StatusOK, nil)
			return
		}
	}
}
