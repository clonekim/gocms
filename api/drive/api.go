package drive

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"gocms/env"
	"google.golang.org/api/drive/v3"
	"strings"
)

func Init( e *echo.Echo) {
	g := e.Group("/api")

	g.GET("/drives", GetDriveInfo)
	g.GET("/drive", GetFilesInDrive)
	g.GET("/drive/:id", GetFileDetails)
	g.POST("/drive", CreateFile)
	g.DELETE("/drive/:id", DeleteFile)

}


func GetDriveInfo (c echo.Context) error {
	cms := c.(*env.GoCms)

	results, err := cms.Drives.Drives.List().Fields("kind, nextPageToken, drives").Do()

	if err != nil {
		return c.JSON(500, err)
	}

	return c.JSON(200, results)
}


/*
  모든 파일 가져오기
  폴더, 파일이든 모드 파일로 취급
 */
func GetFilesInDrive(c echo.Context) error {
	cms := c.(*env.GoCms)

	results, err := cms.Drives.Files.
		List().
		SupportsAllDrives(true).
		Fields("files(id, kind, mimeType, name, description, starred, trashed, parents, owners, webContentLink, webViewLink, iconLink, spaces, thumbnailLink, createdTime, modifiedTime, sharingUser, size, capabilities)").
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
		Fields("id, kind, mimeType").
		Do()


	if err != nil {
		return c.JSON(500, err)
	} else {
		return c.JSON(200, result)
	}
}

type FileRequest struct {
	Name     string `json:"name" validate:"required"`
	MimeType string `json:"mimeType,omitempty" validate:"required"`
	Parents []string `json:"parents,omitempty"`
	Email string `json:"email,omitempty" validate:"email"`
}




func CreateFile( c echo.Context) (err error) {

	var isMultipart = true
	form, err := c.MultipartForm()

	var fileRequest = new(FileRequest)

	if  err != nil {
		c.Logger().Error(err)
		_ = c.Bind(&fileRequest)
		isMultipart = false
	}

	getFirstValue := func( field []string) string {

		if len(field) > 0 {
			return strings.TrimSpace(field[0])
		}
		return ""
	}

	if isMultipart {

		fileRequest = &FileRequest{
			Name: form.Value["name"][0],
			MimeType: getFirstValue(form.Value["mimeType"]),
			Parents: form.Value["parents"],
			Email: getFirstValue(form.Value["email"]),
		}
	}

	err = c.Validate(fileRequest); if err != nil {
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

	if isMultipart && form != nil {

		file := form.File["files"][0:1][0]

		src, err := file.Open()

		if err != nil {
			return err
		}

		defer src.Close()

		result, err := cms.Drives.Files.
		Create(googleFile).
		Media(src).
		ProgressUpdater(func(current, total int64) {
			fmt.Printf("-> %d, %d\r", current, total)

		}).Do()

		if err != nil {
			return c.JSON(500, err)
		}


		apiResult = result

	} else {

		result, err := cms.Drives.Files.Create(googleFile).Do()

		if err != nil {
			return c.JSON(500, err)
		}

		apiResult = result
	}


	if fileRequest.Email != "" {
		err := NewPermission(cms, apiResult.Id, fileRequest.Email)
		if err != nil {
			cms.Drives.Files.Delete(apiResult.Id)
			cms.Logger().Debugf("giving a new permission - %s was failed", fileRequest.Email)
			c.JSON(500, err)
		}
	}

	return c.JSON(200, apiResult)


}


func NewPermission(cms *env.GoCms,  fileId string, email string) error {
	_, err := cms.Drives.Permissions.Create(fileId, &drive.Permission{
		EmailAddress: email,
		Type: "user",
		Role: "reader"}).
		TransferOwnership(false).
		SupportsAllDrives(true).
		SendNotificationEmail(true).Do()

	if err != nil {
		return err
	}

	return nil
}


func DeleteFile( c echo.Context) error {
	id := c.Param("id")
	cms := c.(*env.GoCms)
	err := cms.Drives.Files.Delete(id).Do()

	if err != nil {
		return c.JSON(500, err)
	} else {
		return c.NoContent(200)
	}
}


func ShareFile( c echo.Context) error  {
	id := c.Param("id")
	cms := c.(*env.GoCms)

	result, err := cms.Drives.Permissions.Create(id, &drive.Permission{

		EmailAddress: "clonekim@gmail.com",
		//DisplayName: "clonekim@gmail.com",
		Type: "user",
		Role: "owner",
		//AllowFileDiscovery: true,
		//Domain: "hist.co.kr",
		//AllowFileDiscovery: true,
	}).TransferOwnership(true).SupportsAllDrives(true).SendNotificationEmail(true) .Do()

	if err != nil {
		return c.JSON(500, err)
	} else {
		return c.JSON(200, result)
	}
}



func Upload( c echo.Context) error {

	form, err := c.MultipartForm()

	if  err != nil {
		return err
	}


	file := form.File["files"][0:1][0]
	src, err := file.Open()

	if err != nil {
		return err
	}

	cms := c.(*env.GoCms)
	drivefile := &drive.File{Name: file.Filename}
	result, err := cms.Drives.Files.
		Create(drivefile).
		Media(src).
		ProgressUpdater(func(current, total int64) {
		  fmt.Printf("-> %d, %d\r", current, total)

	}).Do()

	defer src.Close()

	if err != nil {
		return c.JSON(500, err)
	}

	return c.JSON(200, result)

}