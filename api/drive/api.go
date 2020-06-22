package drive

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"gocms/env"
	"google.golang.org/api/drive/v3"
	"strings"
)

func Init(e *echo.Echo) {
	g := e.Group("/api")

	g.GET("/drives", GetDriveInfo)
	g.GET("/drive", GetFilesInDrive)
	g.GET("/drive/:id", GetFileDetails)
	g.POST("/drive", CreateFile)
	g.DELETE("/drive/:id", DeleteFile)

}

func GetDriveInfo(c echo.Context) error {
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

		file := form.File["files"][0:1][0]

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
			cms.Logger().Debugf("giving a new permission - %s was failed", fileRequest.Email)
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
