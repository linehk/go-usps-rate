package zone

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

func TestWriteToChart(t *testing.T) {
	ori := "94010"
	f := excelize.NewFile()
	fileName := "zone.xlsx"
	f.Path = fileName

	now := time.Now()
	err := WriteToChart(ori, f,
		WithName("TestZone"),
		WithDate{&now},
		WithStartHeaderName("Test Zip Code Start"),
		WithEndHeaderName("Test Zip Code End"),
		WithResultHeaderName("TestZone"))
	if err != nil {
		t.Error(err)
	}

	t.Cleanup(func() {
		if err := os.Remove(fileName); err != nil {
			panic(err)
		}
	})
}

func TestGetByChart(t *testing.T) {
	ori := "94010"
	f := excelize.NewFile()
	fileName := "zone.xlsx"
	f.Path = fileName

	now := time.Now()
	err := WriteToChart(ori, f,
		WithName("TestZone"),
		WithDate{&now},
		WithStartHeaderName("Test Zip Code Start"),
		WithEndHeaderName("Test Zip Code End"),
		WithResultHeaderName("TestZone"))
	if err != nil {
		t.Error(err)
	}

	tests := []struct {
		dst              string
		f                *excelize.File
		resultHeaderName string
		expected         int
		expectedErr      error
	}{
		{"00604", f, "TestZone", 8, nil},
		{"387", f, "TestZone", 7, nil},
		{"91300", f, "TestZone", 3, nil},
	}

	for _, tt := range tests {
		actual, actualErr := GetByChart(tt.dst, tt.f, tt.resultHeaderName)
		assert.Equal(t, actualErr, tt.expectedErr)
		assert.Equal(t, actual, tt.expected)
	}

	t.Cleanup(func() {
		if err := os.Remove(fileName); err != nil {
			panic(err)
		}
	})
}

func TestGet(t *testing.T) {
	tests := []struct {
		ori         string
		dst         string
		date        time.Time
		expected    int
		expectedErr error
	}{
		{"94117", "90304", time.Now(), 4, nil},
		{"97030", "90304", time.Now(), 5, nil},
		{"24301", "90304", time.Now(), 8, nil},
	}

	for _, tt := range tests {
		actual, actualErr := Get(tt.ori, tt.dst, tt.date)
		assert.Equal(t, actualErr, tt.expectedErr)
		assert.Equal(t, actual, tt.expected)
	}
}
