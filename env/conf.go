package env

import (
	"github.com/labstack/echo/v4"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/sheets/v4"
)

var Version = "0.0.1"


type GoCms struct {
	echo.Context
	Drives *drive.Service
	Docs *docs.DocumentsService
	Spreads *sheets.SpreadsheetsService
	ServiceAccountEmail *string
	Debug bool
	Port int
}

