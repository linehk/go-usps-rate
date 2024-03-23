package rate

import (
	"strconv"
	"time"

	"github.com/linehk/go-usps-rate/zone"
	"github.com/xuri/excelize/v2"
)

// weightToRow convert likes 1->5, 2->6...
func weightToRow(w int) string {
	gap := 4
	return strconv.FormatInt(int64(w+gap), 10)
}

// zoneToColumn convert likes 1->B, 2->C...
func zoneToColumn(z int) string {
	asciiGap := 48
	gap := 17
	return string(rune(z + asciiGap + gap))
}

// Parcel is the parcel stats.
// Weight unit is lb.
// Length, Width, Height unit is in.
type Parcel struct {
	Weight int
	Length int
	Width  int
	Height int
}

type serviceOptions struct {
	serviceCode string
	carrierCode string
	date        *time.Time
}

type options interface {
	applyTo(opts *serviceOptions)
}

type WithServiceCode string

func (w WithServiceCode) applyTo(opts *serviceOptions) {
	opts.serviceCode = string(w)
}

type WithCarrierCode string

func (w WithCarrierCode) applyTo(opts *serviceOptions) {
	opts.carrierCode = string(w)
}

type WithDate struct {
	date *time.Time
}

func (w WithDate) applyTo(opts *serviceOptions) {
	opts.date = w.date
}

// Get return rate by API.
func Get(ori, dst string, f *excelize.File, parcel Parcel, opts ...options) (float64, error) {
	now := time.Now()
	option := serviceOptions{
		serviceCode: USPS,
		date:        &now,
	}
	for _, opt := range opts {
		opt.applyTo(&option)
	}

	zoneValue, err := zone.Get(ori, dst, *option.date)
	if err != nil {
		return 0, err
	}

	div := divisor(option.serviceCode, option.carrierCode)
	weight := billableWeight(parcel.Weight, dimensionalWeight(parcel.Length, parcel.Width, parcel.Height, div))

	rate, err := getRateByExcel(zoneValue, weight, f)
	if err != nil {
		return 0, err
	}

	return rate, nil
}

func getRateByExcel(zone, weight int, f *excelize.File) (float64, error) {
	rateStr, err := f.GetCellValue("Formula", zoneToColumn(zone)+weightToRow(weight))
	if err != nil {
		return 0, err
	}
	rate, err := strconv.ParseFloat(rateStr, 64)
	if err != nil {
		return 0, err
	}
	return rate, nil
}

func dimensionalWeight(l, w, h, div int) int {
	return (l * w * h) / div
}

func billableWeight(actualWeight, dimensionalWeight int) int {
	if dimensionalWeight >= actualWeight {
		return dimensionalWeight
	}
	return actualWeight
}

// ServiceCode
const (
	USPS  = "usps"
	FedEx = "fedex"
	UPS   = "ups"
)

// CarrierCode
const (
	UPSDailyPackages  = "Daily Packages"
	UPSRetailPackages = "Retail Packages"
)

func divisor(serviceCode string, carrierCode string) int {
	if serviceCode == USPS {
		return 166
	}
	if serviceCode == FedEx {
		return 139
	}
	if serviceCode == UPS {
		if carrierCode == UPSDailyPackages {
			return 139
		}
		if carrierCode == UPSRetailPackages {
			return 166
		}
	}
	return 166
}
