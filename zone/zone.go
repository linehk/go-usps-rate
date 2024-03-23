package zone

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/xuri/excelize/v2"
)

type singular struct {
	OriginError       string `json:"OriginError,omitempty"`
	DestinationError  string `json:"DestinationError,omitempty"`
	ShippingDateError string `json:"ShippingDateError,omitempty"`
	PageError         string `json:"PageError,omitempty"`
	EffectiveDate     string `json:"EffectiveDate,omitempty"`
	ZoneInformation   string `json:"ZoneInformation,omitempty"`
}

// Get return shipping zone by API.
// ori, dst must be 5 digit zip code.
func Get(ori, dst string, date time.Time) (int, error) {
	layout := "01/02/2006"
	shippingDate := date.Format(layout)
	endpoint := fmt.Sprintf(
		"https://postcalc.usps.com/DomesticZoneChart/GetZone?origin=%s&destination=%s&shippingDate=%s",
		ori, dst, shippingDate)
	resp, err := http.Get(endpoint)
	if err != nil {
		return 0, err
	}

	zoneSingular := singular{}
	err = json.NewDecoder(resp.Body).Decode(&zoneSingular)
	if err != nil {
		return 0, err
	}

	// ZoneInformation string like: `The Zone is 4. This is not a Local Zone. The destination ZIP Code is not within the same NDC as the origin ZIP Code.`
	// so we need the 12th character.
	zoneCharIndex := 12
	zoneRaw := string(rune(zoneSingular.ZoneInformation[zoneCharIndex]))
	zone, err := strconv.Atoi(zoneRaw)
	if err != nil {
		return 0, err
	}

	return zone, nil
}

type chart struct {
	ZIPCodeError      string        `json:"ZIPCodeError,omitempty"`
	ShippingDateError string        `json:"ShippingDateError,omitempty"`
	PageError         string        `json:"PageError,omitempty"`
	EffectiveDate     string        `json:"EffectiveDate,omitempty"`
	Column0           []chartColumn `json:"Column0,omitempty"`
	Column1           []chartColumn `json:"Column1,omitempty"`
	Column2           []chartColumn `json:"Column2,omitempty"`
	Column3           []chartColumn `json:"Column3,omitempty"`
	Zip5Digit         []chartColumn `json:"Zip5Digit,omitempty"`
}

type chartColumn struct {
	ZipCodes    string `json:"ZipCodes,omitempty"`
	Zone        string `json:"Zone,omitempty"`
	MailService string `json:"MailService,omitempty"`
}

type sheetOptions struct {
	name             string
	date             *time.Time
	startHeaderName  string
	endHeaderName    string
	resultHeaderName string
}

type options interface {
	applyTo(opts *sheetOptions)
}

type WithName string

func (w WithName) applyTo(opts *sheetOptions) {
	opts.name = string(w)
}

type WithDate struct {
	date *time.Time
}

func (w WithDate) applyTo(opts *sheetOptions) {
	opts.date = w.date
}

type WithStartHeaderName string

func (w WithStartHeaderName) applyTo(opts *sheetOptions) {
	opts.startHeaderName = string(w)
}

type WithEndHeaderName string

func (w WithEndHeaderName) applyTo(opts *sheetOptions) {
	opts.endHeaderName = string(w)
}

type WithResultHeaderName string

func (w WithResultHeaderName) applyTo(opts *sheetOptions) {
	opts.resultHeaderName = string(w)
}

// WriteToChart write zone chart to excel.
func WriteToChart(ori string, f *excelize.File, opts ...options) error {
	now := time.Now()
	option := sheetOptions{
		name:             "Zone",
		date:             &now,
		startHeaderName:  "Zip Code Start",
		endHeaderName:    "Zip Code End",
		resultHeaderName: "Zone",
	}
	for _, opt := range opts {
		opt.applyTo(&option)
	}

	layout := "01/02/2006"
	shippingDate := option.date.Format(layout)
	if len(ori) == 5 {
		ori = ori[0:3]
	}
	endpoint := fmt.Sprintf(
		"https://postcalc.usps.com/DomesticZoneChart/GetZoneChart?zipCode3Digit=%s&shippingDate=%s",
		ori, shippingDate)
	resp, err := http.Get(endpoint)
	if err != nil {
		return err
	}

	zoneChart := chart{}
	err = json.NewDecoder(resp.Body).Decode(&zoneChart)
	if err != nil {
		return err
	}

	sheetIndex, err := f.NewSheet(option.name)
	if err != nil {
		return err
	}

	// write header
	err = f.SetCellValue(option.name, "A1", option.startHeaderName)
	if err != nil {
		return err
	}
	err = f.SetCellValue(option.name, "B1", option.endHeaderName)
	if err != nil {
		return err
	}
	err = f.SetCellValue(option.name, "C1", option.resultHeaderName)
	if err != nil {
		return err
	}

	// 2 is for skip header
	index := 2
	index, err = setCellValueByColumn(zoneChart.Column0, f, index, option.name)
	if err != nil {
		return err
	}
	index, err = setCellValueByColumn(zoneChart.Column1, f, index, option.name)
	if err != nil {
		return err
	}
	index, err = setCellValueByColumn(zoneChart.Column2, f, index, option.name)
	if err != nil {
		return err
	}
	_, err = setCellValueByColumn(zoneChart.Column3, f, index, option.name)
	if err != nil {
		return err
	}

	f.SetActiveSheet(sheetIndex)
	err = f.Save()
	if err != nil {
		return err
	}

	return nil
}

func setCellValueByColumn(columns []chartColumn, f *excelize.File, i int, sheetName string) (int, error) {
	for _, column := range columns {
		zipCodeStartValue := ""
		zipCodeEndValue := ""

		// ZipCodes format like: 270
		if len(column.ZipCodes) == 3 {
			zipCodeStartValue = column.ZipCodes
			zipCodeEndValue = column.ZipCodes
		}

		// ZipCodes format like: 270---342
		if len(column.ZipCodes) > 3 {
			zipCodeStartValue = column.ZipCodes[0:3]
			zipCodeEndValue = column.ZipCodes[6:9]
		}

		zipCodeStartCell, err := excelize.CoordinatesToCellName(1, i)
		if err != nil {
			return 0, err
		}
		err = f.SetCellValue(sheetName, zipCodeStartCell, zipCodeStartValue)
		if err != nil {
			return 0, err
		}

		zipCodeEndCell, err := excelize.CoordinatesToCellName(2, i)
		if err != nil {
			return 0, err
		}
		err = f.SetCellValue(sheetName, zipCodeEndCell, zipCodeEndValue)
		if err != nil {
			return 0, err
		}

		zoneCell, err := excelize.CoordinatesToCellName(3, i)
		if err != nil {
			return 0, err
		}
		err = f.SetCellValue(sheetName, zoneCell, string(column.Zone[0]))
		if err != nil {
			return 0, err
		}

		i++
	}

	return i, nil
}

// GetByChart return shipping zone by excel.
func GetByChart(dst string, f *excelize.File, resultHeaderName string) (int, error) {
	rows, err := f.GetRows(resultHeaderName)
	if err != nil {
		return 0, err
	}
	if len(dst) == 5 {
		dst = dst[0:3]
	}
	dstNum, err := strconv.Atoi(dst)
	if err != nil {
		return 0, err
	}

	// skip header
	headerIndex := 1
	for _, row := range rows[headerIndex:] {
		zipCodeStart, convertErr := strconv.Atoi(row[0])
		if convertErr != nil {
			return 0, convertErr
		}

		zipCodeEnd, convertErr := strconv.Atoi(row[1])
		if convertErr != nil {
			return 0, convertErr
		}

		zone, convertErr := strconv.Atoi(row[2])
		if convertErr != nil {
			return 0, convertErr
		}

		if dstNum >= zipCodeStart && dstNum <= zipCodeEnd {
			return zone, nil
		}
	}

	return 0, errors.New("cannot found zone")
}
