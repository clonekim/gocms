package cmd

import (
	"context"
	"errors"
	"gocms/env"
	"gocms/server"
	"io/ioutil"
	"net/http"

	"github.com/urfave/cli/v2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

var globalContext = context.Background()

func Setup() *cli.App {

	return &cli.App{
		Name:     "gocms " + env.Version,
		Usage:    "google cms implemented by Go.",
		HideHelp: false,
		Flags: []cli.Flag{

		/*	&cli.BoolFlag{
				Name:  "debug",
				Value: false,
				Usage: "debug option",
			},*/

			&cli.IntFlag{
				Name:  "port",
				Value: 8000,
				Usage: "port number",
			},

			&cli.StringFlag{
				Name:  "credential",
				Value: "",
				Usage: "credential file to authenticate",
			},
		},

		Action: func(c *cli.Context) error {
			c.App.Setup()

			port := c.Int("port")
		//	debug := c.Bool("debug")

			var gocms = env.GoCms{}

			if c.String("credential") == "" {
				return errors.New("credential file is required")
			}

			b, err := ioutil.ReadFile(c.String("credential"))
			if err != nil {
				return err
			}

			conf, err := createJWTConfig(b)
			if err != nil {
				return err
			}

			client := conf.Client(globalContext)

			drives, err := createDriveService(client)
			if err != nil {
				return err
			}

			docs, err := createDocumentService(client)
			if err != nil {
				return err
			}

			sheet, err := createSheetService(client)
			if err != nil {
				return err
			}

			gocms.ServiceAccountEmail = &conf.Email
			gocms.Drives = drives
			gocms.Spreads = sheet.Spreadsheets
			gocms.Docs = docs.Documents

			gocms.Port = port
			//gocms.Debug = debug

			server.StartHTTPD(&gocms)

			return nil

		},
	}

}

func createJWTConfig(data []byte) (*jwt.Config, error) {
	//log.Println("Creating HttpClient")

	//*jwt.Config
	conf, err := google.JWTConfigFromJSON(data,
		drive.DriveScope,
		drive.DriveFileScope,
		drive.DriveAppdataScope,
		drive.DriveScriptsScope,
		drive.DriveMetadataScope,
		drive.DriveMetadataReadonlyScope,
		drive.DrivePhotosReadonlyScope,
		docs.DocumentsScope,
		docs.DocumentsReadonlyScope,
		sheets.SpreadsheetsScope,
		sheets.SpreadsheetsReadonlyScope,
	)

	if err != nil {
		return nil, err
	}

	return conf, err
}

func createDriveService(client *http.Client) (*drive.Service, error) {
	//log.Println("Creating DriveService")
	return drive.NewService(globalContext, option.WithHTTPClient(client))
}

func createSheetService(client *http.Client) (*sheets.Service, error) {
	//log.Println("Creating SpreadService")
	return sheets.NewService(globalContext, option.WithHTTPClient(client))
}

func createDocumentService(client *http.Client) (*docs.Service, error) {
	//log.Println("Creating DocumentService")
	return docs.NewService(globalContext, option.WithHTTPClient(client))
}
