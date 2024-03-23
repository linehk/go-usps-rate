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
		expected    float64
		expectedErr error
	}{
		{"94117", "90304", f, parcel1, 13.63, nil},
		{"97030", "90304", f, parcel2, 14.09, nil},
		{"24301", "90304", f, parcel3, 285.78, nil},
	}

	for _, tt := range tests {
		actual, actualErr := Get(tt.ori, tt.dst, tt.f, tt.parcel,
			WithServiceCode(USPS),
			WithCarrierCode(UPSDailyPackages),
			WithDate{&now})
		assert.Equal(t, actualErr, tt.expectedErr)
		assert.Equal(t, actual, tt.expected)
	}
}
