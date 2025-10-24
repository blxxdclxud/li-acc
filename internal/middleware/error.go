package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		err := c.Errors.Last()
		if err == nil {
			return
		}

		message := Localizer(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": message,
		})
		c.Abort()
	}
}
