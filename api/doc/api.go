package doc

import (
	"github.com/labstack/echo/v4"
	"gocms/env"
	"google.golang.org/api/docs/v1"
)



func Init( e *echo.Echo) {
	g := e.Group("/api")
	g.POST("/doc", CreateDocument)
	g.GET("/doc/:id", GetDocument)
	g.PUT("/doc/:id", BatchUpdateDocument)
}


type DocumentRequest struct {
	Title string `json:"title"`
}


func CreateDocument(c echo.Context) error {

	docRequest := new(DocumentRequest)

	err := c.Bind(docRequest); if err != nil {
		return err
	}

	cms := c.(*env.GoCms)
	doc := &docs.Document{ Title: docRequest.Title }
	result, err := cms.Docs.Create(doc).Do()

	if err != nil {
		return c.JSON(500, err)
	}

	return c.JSON(200, result)
}


func GetDocument ( c echo.Context) error {
	cms := c.(*env.GoCms)
	docId := c.Param("id")

	res, err := cms.Docs.Get(docId).Do()

	if err != nil {
		return c.JSON(500, err)
	}

	return c.JSON(200, res)
}


func BatchUpdateDocument ( c echo.Context) error {
	cms := c.(*env.GoCms)
	docId := c.Param("id")
	var requests = make([]*docs.Request, 0)

	err := c.Bind(requests); if err != nil {
		return err
	}

	res, err := cms.Docs.BatchUpdate(docId, &docs.BatchUpdateDocumentRequest{Requests: requests}).Do()

	if err != nil {
		return c.JSON(500, err)
	}

	return c.JSON(200, res)
}
