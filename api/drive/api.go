package drive

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"gocms/env"
	"google.golang.org/api/drive/v3"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

func Init(e *echo.Echo) {
	g := e.Group("/api")

	g.GET("/drives", GetDriveInfo)
	g.GET("/drive", GetFilesInDrive)
	g.GET("/drive/:id", GetFileDetails)
	g.GET("/download/:id", Download)
	g.POST("/drive", CreateFile)
	g.DELETE("/drive/:id", DeleteFile)

}

func GetDriveInfo(c echo.Context) error {
	cms := c.(*env.GoCms)

	results, err := cms.Drives.Files.Get("root").Do()

	if err != nil {
		return c.JSON(500, err)
	}

	return c.JSON(200, results)
}


func getPageSize(param string) int64 {
	if param != "" {
		i, err := strconv.ParseInt(param, 10, 64)
		if err != nil {
			return 20
		}
		return i
	}

	return 20
}



/*
  모든 파일 가져오기
  폴더, 파일이든 모드 파일로 취급
*/
func GetFilesInDrive(c echo.Context) error {
	cms := c.(*env.GoCms)



	results, err := cms.Drives.Files.
		List().
		//DriveId("root").
		//SupportsAllDrives(true).
		//IncludeItemsFromAllDrives(true).
		Corpora("user").
	//	Fields("files(id, kind, mimeType, name, description, starred, trashed, parents, owners, webContentLink, webViewLink, iconLink, spaces, thumbnailLink, createdTime, modifiedTime, sharingUser, size, capabilities),  nextPageToken").
		Fields("files(id, kind, mimeType, name, description, starred, trashed, parents, owners,  webViewLink, iconLink, spaces, createdTime, modifiedTime, sharingUser, size), nextPageToken").
		PageSize(getPageSize(c.QueryParam("pageSize"))).
		Q(c.QueryParam("q")).

		OrderBy(c.QueryParam("orderby")).
		PageToken( c.QueryParam("pageToken")).
		Do()

	if err != nil {
		return c.JSON(500, err)
	} else {
		return c.JSON(200, results)
	}

}

func GetFileDetails(c echo.Context) error {
	id := c.Param("id")
	cms := c.(*env.GoCms)
	result, err := cms.Drives.Files.
		Get(id).
		SupportsAllDrives(true).
		Fields("id, kind, mimeType, name, description, starred, trashed, owners, webViewLink, iconLink, spaces, createdTime, modifiedTime, sharingUser, size, capabilities").
		Do()

	if err != nil {
		return c.JSON(500, err)
	} else {
		return c.JSON(200, result)
	}
}

type FileRequest struct {
	Name     string   `json:"name" validate:"required"`
	MimeType string   `json:"mimeType,omitempty" validate:"required"`
	Parents  []string `json:"parents,omitempty"`
	Email    string   `json:"email,omitempty" validate:"email"`
}



func CreateFile(c echo.Context) (err error) {

	form, err := c.MultipartForm()

	var fileRequest = new(FileRequest)

	if err != nil { //form type이 multipart가 아니면 에러 발생
		err := c.Bind(&fileRequest)
		if err != nil {
			return err
		}
	} else {

		getFirstValue := func(field []string) string {

			if len(field) > 0 {
				return strings.TrimSpace(field[0])
			}
			return ""
		}

		fileRequest = &FileRequest{
			Name:     form.Value["name"][0],
			MimeType: getFirstValue(form.Value["mimeType"]),
			Parents:  form.Value["parents"],
			Email:    getFirstValue(form.Value["email"]),
		}
	}

	err = c.Validate(fileRequest)
	if err != nil {
		return c.JSON(400, err)
	}

	cms := c.(*env.GoCms)

	googleFile := &drive.File{
		Name: fileRequest.Name,
	}

	if fileRequest.MimeType != "" {
		googleFile.MimeType = fileRequest.MimeType
	}

	if len(fileRequest.Parents) > 0 {
		googleFile.Parents = fileRequest.Parents
	}

	var apiResult *drive.File
	var apiErr error

	if form != nil {

		file := form.File["file"][0:1][0]

		src, err := file.Open()

		if err != nil {
			return err
		}

		defer src.Close()

		apiResult, apiErr = cms.Drives.Files.
			Create(googleFile).
			Media(src).
			ProgressUpdater(func(current, total int64) {
				fmt.Printf("UPLOAD PROGRESS -> %d, %d\r", current, total) //TODO 콜백처리
			}).Do()

	} else {
		apiResult, apiErr = cms.Drives.Files.Create(googleFile).Do()
	}

	if apiErr != nil {
		return c.JSON(500, apiErr)
	}

	if fileRequest.Email != "" {
		_, err := NewPermission(cms, apiResult.Id, fileRequest.Email, "owner")
		if err != nil {
			cms.Drives.Files.Delete(apiResult.Id)
			logrus.Errorf("giving a new permission - %s was failed", fileRequest.Email)
			return c.JSON(500, err)
		}
	}

	return c.JSON(200, apiResult)

}



func NewPermission(cms *env.GoCms, fileId string, email string, role string) (*drive.Permission, error) {
	result, err := cms.Drives.Permissions.Create(fileId, &drive.Permission{
		EmailAddress: email,
		DisplayName:  email,
		Type:         "user",
		Role:         role}).
		TransferOwnership(func() bool {
			if role == "owner" {
				return true
			}
			return false
		}()).
		SupportsAllDrives(true).
		SendNotificationEmail(true).Do()

	if err != nil {
		return nil, err
	}

	return result, nil
}

func DeleteFile(c echo.Context) error {
	id := c.Param("id")
	cms := c.(*env.GoCms)
	err := cms.Drives.Files.Delete(id).Do()

	if err != nil {
		return c.JSON(500, err)
	} else {
		return c.NoContent(200)
	}
}


func Download( c echo.Context) error {
	id := c.Param("id")
	mimetype := "application/octet-stream"
	//mimetype := c.QueryParam("mimeType")
	cms := c.(*env.GoCms)

	//if mimetype == "" {
	//	logrus.Debug("mimeType is empty")
		file, _ := cms.Drives.Files.Get(id).Do()
		mimetype =  file.MimeType
	//	logrus.Debugf("reset mimeType -> %s\n", mimetype)
	//}
	//
	//if mimetype == "" {
	//	mimetype = "application/octet-stream"
	//}

	var res *http.Response
	var err error

	//if mimetype == "application/octet-stream" {
		res, err = cms.Drives.Files.Get(id).Download()
	//} else {
	//	res, err = cms.Drives.Files.Export(id, mimetype).Download()
	//}

	if err != nil {
		logrus.Error(err)
		return c.JSON(500, err)
	}

	contents, err := ioutil.ReadAll(res.Body)

	if err != nil {
		logrus.Error(err)
		return  c.JSON(500, err)
	}

	defer res.Body.Close()

	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("%s; filename=%q", "attachment",  file.Name))
	return c.Blob(200, mimetype , contents)

}


func DeletePermission(cms *env.GoCms, fileId string, permId string) {

	cms.Drives.Permissions.Delete(fileId, permId).Do()
}

func updatePermission(cms *env.GoCms, fileId string, permId string) {
	cms.Drives.Permissions.Update(fileId, permId, &drive.Permission{
		Role: "",
		EmailAddress: "",
		DisplayName: "",

	})
}