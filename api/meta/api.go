package meta

import (
	"github.com/labstack/echo/v4"
	"gocms/env"
)


func Init( e *echo.Echo) {
	e.GET("/meta", GetMeta)
}

func GetMeta(c echo.Context) error {
	cms := c.(*env.GoCms)

	return c.JSON(200, map[string]interface{}{
		"Version":             env.Version,
		"ServiceAccountEmail": cms.ServiceAccountEmail,
	})
}

