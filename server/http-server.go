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
	"log"
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
		fmt.Printf("Errors => %v\n", v.Errors)
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
		"Message": "Hello,",
		"uuid":   uuid.NewV4().String(),
	})
}


func assetHandler() http.Handler {
	box := rice.MustFindBox("../public")
	return http.FileServer(box.HTTPBox())
}


func templateHandler() *rice.Box {
	box, err := rice.FindBox("../public/templates")

	if err != nil {
		log.Fatal(err)
	}

	return box
}


func StartHTTPD(cms *env.GoCms) {
	fmt.Println("Starting GoCms")

	e := echo.New()
	e.HideBanner = true

	assetHandler := assetHandler()

	logLevel := func() logrus.Level {
		if cms.Debug {
			return logrus.DebugLevel
		}
		return logrus.InfoLevel
	}()


	logrus.SetLevel(logLevel)
	logrus.SetFormatter(&logrus.TextFormatter{})

	e.Validator = &CustomValidator{}
	e.Pre(middleware.RemoveTrailingSlash())

	e.Renderer = &Template{ box:  templateHandler() }
	/*
	  로그 포맷 참고
	 https://echo.labstack.com/middleware/logger
	*/
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${time_rfc3339} ${method} - ${uri} ${status}\n",
	}))

	e.GET("/", echo.WrapHandler(assetHandler))
	e.GET("/*", echo.WrapHandler(http.StripPrefix("/", assetHandler)))
	e.GET("/html/:name", func(c echo.Context) error {
		return c.Render(200, c.Param("name"), nil)
	})

	e.Use(middleware.Recover())

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

	meta.Init(e)
	drive.Init(e)
	doc.Init(e)


	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", cms.Port)))
}

