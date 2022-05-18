package main

import (
	"errors"
	"fmt"
	"log"
	"math"
	"os"

	"github.com/google/tiff"
	_ "github.com/google/tiff/bigtiff"
	_ "github.com/google/tiff/geotiff"
)

const (
	ModelPixelScaleTagID = 33550
	ModelTiePointTagID   = 33922
)

func main() {

	fmt.Println("Hi from go!")

	f, err := os.Open("aus_ppp_2020_constrained.tif")
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer f.Close()

	t, err := tiff.Parse(f, nil, nil)
	if err != nil {
		log.Fatalf("Error parsing tiff: %v", err)
	}

	fields := t.IFDs()[0].Fields()

	for _, field := range fields {
		fmt.Printf("%s: %+v\n", field.Tag().Name(), field)
	}

	tiePoints, err := TiePoints(t)
	if err != nil {
		log.Fatalf("Error getting tie points: %v", err)
	}
	fmt.Printf("%+v", tiePoints)

	scale, err := PixelScale(t)
	if err != nil {
		log.Fatalf("Error getting pixel scale: %v", err)
	}
	fmt.Printf("%+v", scale)
}

type TiePoint struct {
	I, J, K float64
	X, Y, Z float64
}

func TiePoints(t tiff.TIFF) ([]TiePoint, error) {
	for _, ifd := range t.IFDs() {
		if ifd.HasField(ModelTiePointTagID) {
			points := t.IFDs()[0].GetField(ModelTiePointTagID)

			if points.Count()%6 != 0 {
				return nil, errors.New("unexpected count of ModelTiePointTag")
			}

			values := make([]float64, points.Count())
			b := points.Value().Bytes()

			for i := 0; i < len(values); i++ {
				bits := points.Value().Order().Uint64(b[i*8 : (i+1)*8])
				values[i] = math.Float64frombits(bits)
			}

			tiePoints := make([]TiePoint, len(values)/6)
			for i := 0; i < len(tiePoints); i++ {
				tiePoints[i] = TiePoint{
					I: values[i*6+0],
					J: values[i*6+1],
					K: values[i*6+2],
					X: values[i*6+3],
					Y: values[i*6+4],
					Z: values[i*6+5],
				}
			}
			return tiePoints, nil
		}
	}
	return nil, errors.New("ModelTiePointTag not found")
}

type ScaleXYZ struct {
	X, Y, Z float64
}

func PixelScale(t tiff.TIFF) (*ScaleXYZ, error) {
	for _, ifd := range t.IFDs() {
		if ifd.HasField(ModelPixelScaleTagID) {
			s := t.IFDs()[0].GetField(ModelPixelScaleTagID)

			if s.Count() != 3 {
				return nil, errors.New("unexpected count of ModelPixelScaleTag")
			}

			b := s.Value().Bytes()

			scale := ScaleXYZ{
				X: math.Float64frombits(s.Value().Order().Uint64(b[0*8 : (0+1)*8])),
				Y: math.Float64frombits(s.Value().Order().Uint64(b[1*8 : (2+1)*8])),
				Z: math.Float64frombits(s.Value().Order().Uint64(b[2*8 : (2+1)*8])),
			}

			return &scale, nil
		}
	}
	return nil, errors.New("ModelPixelScaleTag not found")
}
