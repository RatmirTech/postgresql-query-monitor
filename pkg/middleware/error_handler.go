package middleware

import (
	"net/http"

	"github.com/dreadew/go-common/pkg/errors"

	"github.com/gin-gonic/gin"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			}
		}()

		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			switch e := err.(type) {
			case *errors.ValidationError:
				c.JSON(http.StatusBadRequest, gin.H{
					"title": "Ошибка валидации",
					"error": e.Error(),
				})
			case *errors.UserVisibleError:
				c.JSON(http.StatusBadRequest, gin.H{
					"title": "Ошибка!",
					"error": e.Details,
				})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "internal server error",
				})
			}
		}

		c.Abort()
	}
}
