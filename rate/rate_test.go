package rate

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

func TestGet(t *testing.T) {
	f, err := excelize.OpenFile("rate.xlsx")
	if err != nil {
		t.Error(err)
	}

	parcel1 := Parcel{
		Weight: 1,
		Length: 0,
		Width:  0,
		Height: 0,
	}
	parcel2 := Parcel{
		Weight: 1,
		Length: 0,
		Width:  0,
		Height: 0,
	}
	parcel3 := Parcel{
		Weight: 1,
		Length: 20,
		Width:  20,
		Height: 20,
	}

	now := time.Now()

	tests := []struct {
		ori         string
		dst         string
		f           *excelize.File
		parcel      Parcel
		opts        []options
		expected    float64
		expectedErr error
	}{
		{"94117", "90304", f, parcel1,
			[]options{
				WithServiceCode(USPS),
				WithDate{&now}},
			13.63, nil},
		{"97030", "90304", f, parcel2,
			[]options{
				WithServiceCode(FedEx),
				WithDate{&now}},
			14.09,
			nil},
		{"24301", "90304", f, parcel3,
			[]options{
				WithServiceCode(UPS),
				WithCarrierCode(UPSDailyPackages),
				WithDate{&now}},
			320.84,
			nil},
		{"24301", "90304", f, parcel3,
			[]options{
				WithServiceCode(UPS),
				WithCarrierCode(UPSRetailPackages),
				WithDate{&now}},
			285.78,
			nil},
		{"24301", "90304", f, parcel3,
			[]options{},
			285.78,
			nil},
	}

	for _, tt := range tests {
		actual, actualErr := Get(tt.ori, tt.dst, tt.f, tt.parcel, tt.opts...)
		assert.Equal(t, actualErr, tt.expectedErr)
		assert.Equal(t, actual, tt.expected)
	}
}
