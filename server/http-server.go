package server

import (
	"fmt"
	"github.com/gookit/validate"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	"gocms/api/doc"
	"gocms/api/drive"
	"gocms/api/meta"
	"gocms/env"
	"github.com/GeertJohan/go.rice"
	"html/template"
	"io"
	"net/http"
	"github.com/satori/go.uuid"
)


type CustomValidator struct {
}

func (c *CustomValidator) Validate(i interface{}) error {
	v := validate.Struct(i)

	if v.Validate() {
		return nil
	} else {
		logrus.Errorf("Errors => %v\n", v.Errors)
		return v.Errors
	}
}

type Template struct {
	box *rice.Box
}


func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	tmplString, err := t.box.String(name);  if err != nil {
		return err
	}

	tmplMessage, err := template.New(name).Parse(tmplString); if err != nil {
		return err
	}

	return tmplMessage.Execute(w, map[string]interface{}{
		"serviceAccountEmail": goCms.ServiceAccountEmail,
		"uuid":   uuid.NewV4().String(),
	})
}


var goCms *env.GoCms




func StartHTTPD(cms *env.GoCms) {
	fmt.Println("Starting GoCms")
	goCms = cms


	e := echo.New()
	e.HideBanner = true
	e.Debug = cms.Debug
	e.Logger = Logger{ logrus.StandardLogger(),}


	e.Use(LoggerHook())
	e.Use(middleware.RequestID())
	e.Use(middleware.Recover())

	e.Validator = &CustomValidator{}
	e.Pre(middleware.RemoveTrailingSlash())


	/*
	  로그 포맷 참고
	 https://echo.labstack.com/middleware/logger
	*/
	//e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
	//	Format: "${time_rfc3339} ${method} - ${uri} ${status}\n",
	//}))


	e.GET("/html/:name", func(c echo.Context) error {
		return c.Render(200, c.Param("name"), nil)
	})


	//custom middleware
	//https://echo.labstack.com/guide/context
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cmsCtx := &env.GoCms{}

			cmsCtx.Context = c
			cmsCtx.ServiceAccountEmail = cms.ServiceAccountEmail
			cmsCtx.Spreads = cms.Spreads
			cmsCtx.Drives = cms.Drives
			cmsCtx.Docs = cms.Docs
			cmsCtx.Debug = cms.Debug

			return next(cmsCtx)
		}
	})

	SetupAsset(e)
	SetupRenderer(e)
	meta.Init(e)
	drive.Init(e)
	doc.Init(e)

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", cms.Port)))
}

func SetupAsset(e *echo.Echo) {
	assetHandler := http.FileServer(rice.MustFindBox("../public").HTTPBox())
	e.GET("/", echo.WrapHandler(assetHandler))
	e.GET("/*", echo.WrapHandler(http.StripPrefix("/", assetHandler)))

}

func SetupRenderer(e *echo.Echo) {
	box, err := rice.FindBox("../public/templates")

	if err != nil {
		logrus.Debug(err)
	}

	e.Renderer = &Template{ box:  box }
}

