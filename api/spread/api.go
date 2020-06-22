package spread

import (
	"github.com/labstack/echo/v4"
)

type SpreadData struct {
	Datas [][] interface{} `json:"datas"`
}


func CreateSpredSheet(c echo.Context) error {

	/*datas := new(SpreadData)

	err := c.Bind(datas)
	if err != nil {
		return err
	}

	cms := c.(*env.GoCms)

	spreadsheetId := "### spreadsheet ID ###"
	//rangeData := "sheet1!A1:B3"
	values := [][]interface{}{{"sample_A1", "sample_B1"}, {"sample_A2", "sample_B2"}, {"sample_A3", "sample_A3"}}
	rb := &sheets.BatchUpdateValuesRequest{
		ValueInputOption: "USER_ENTERED",
	}
	rb.Data = append(rb.Data, &sheets.ValueRange{
	//	Range:  rangeData,
		Values: values,
	})
	_, err = cms.SheetServ.Values.BatchUpdate(spreadsheetId, rb).Do()

	if err != nil {
		return err
	}*/

	return nil
	//
	//gridData := make([]sheets.GridData, len(datas.Datas))
	//
	//for i, data := range datas.Datas {
	//	append(gridData, data[i])
	//}
	//
	//sheet := &sheets.Spreadsheet{
	//	Sheets: &sheets.Sheet{
	//		Data: make(sheets.GridData, len(datas)),
	//	}
	//}
	//
	//result, err := cms.SheetServ.Create(sheet).Do()
	//
	//if err != nil {
	//	return c.JSON(500, err)
	//}
	//return c.JSON(200, result)
}
