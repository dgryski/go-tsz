package main

import (
	"fmt"
	"github.com/dgryski/go-tsz"
	"github.com/dgryski/go-tsz/testdata"
	"math"
	"math/rand"
	"os"
	"text/tabwriter"
)

// collection of 24h worth of minutely points, with different characteristics.
var ConstantZero = make([]tsz.Point, 60*24)
var ConstantOne = make([]tsz.Point, 60*24)
var Batch100ZeroOne = make([]tsz.Point, 60*24)
var FlappingZeroOne = make([]tsz.Point, 60*24)
var ConstantPos3f = make([]tsz.Point, 60*24)
var ConstantNeg3f = make([]tsz.Point, 60*24)
var ConstantPos0f = make([]tsz.Point, 60*24)
var ConstantNeg0f = make([]tsz.Point, 60*24)
var ConstantNearMaxf = make([]tsz.Point, 60*24)
var ConstantNearMinf = make([]tsz.Point, 60*24)
var ConstantNearMax0f = make([]tsz.Point, 60*24)
var ConstantNearMin0f = make([]tsz.Point, 60*24)
var SmallRangePosd = make([]tsz.Point, 60*24)
var LargeRangePosd = make([]tsz.Point, 60*24)
var SmallRangePosf = make([]tsz.Point, 60*24)
var LargeRangePosf = make([]tsz.Point, 60*24)
var SmallRangePos0f = make([]tsz.Point, 60*24)
var LargeRangePos0f = make([]tsz.Point, 60*24)
var RandomSmallPosf = make([]tsz.Point, 60*24)
var RandomSmallf = make([]tsz.Point, 60*24)
var RandomLargePosf = make([]tsz.Point, 60*24)
var RandomLargef = make([]tsz.Point, 60*24)
var RandomLargePos0f = make([]tsz.Point, 60*24)
var RandomLarge0f = make([]tsz.Point, 60*24)

func init() {
	for i := 0; i < len(DataZeroFloats); i++ {
		DataZeroFloats[i] = tsz.Point{float64(0), uint32(i * 60)}
	}
	for i := 0; i < len(DataSameFloats); i++ {
		DataSameFloats[i] = tsz.Point{float64(1234.567), uint32(i * 60)}
	}
	for i := 0; i < len(DataSmallRangePosInts); i++ {
		DataSmallRangePosInts[i] = tsz.Point{testdata.TwoHoursData[i%120].V, uint32(i * 60)}
	}
	for i := 0; i < len(DataSmallRangePosFloats); i++ {
		v := float64(testdata.TwoHoursData[i%120].V) * 1.2
		DataSmallRangePosFloats[i] = tsz.Point{v, uint32(i * 60)}
	}
	for i := 0; i < len(DataLargeRangePosFloats); i++ {
		v := (float64(testdata.TwoHoursData[i%120].V) / 1000) * math.MaxFloat64
		DataLargeRangePosFloats[i] = tsz.Point{v, uint32(i * 60)}
	}
	for i := 0; i < len(DataRandomPosFloats); i++ {
		DataRandomPosFloats[i] = tsz.Point{rand.ExpFloat64(), uint32(i * 60)}
	}
	for i := 0; i < len(DataRandomFloats); i++ {
		DataRandomFloats[i] = tsz.Point{rand.NormFloat64(), uint32(i * 60)}
	}
}
func main() {
	for i := 0; i < 60*24; i++ {
		ts := uint32(i * 60)
		ConstantZero[i] = tsz.Point{float64(0), ts}
		ConstantOne[i] = tsz.Point{float64(1), ts}
		if i%200 < 100 {
			Batch100ZeroOne[i] = tsz.Point{float64(0), ts}
		} else {
			Batch100ZeroOne[i] = tsz.Point{float64(1), ts}
		}
		if i%2 == 0 {
			FlappingZeroOne[i] = tsz.Point{float64(0), ts}
		} else {
			FlappingZeroOne[i] = tsz.Point{float64(1), ts}
		}
		ConstantPos3f[i] = tsz.Point{float64(1234.567), ts}
		ConstantNeg3f[i] = tsz.Point{float64(-1234.567), ts}
		ConstantPos0f[i] = tsz.Point{float64(1234), ts}
		ConstantNeg0f[i] = tsz.Point{float64(-1235), ts}
		ConstantNearMaxf[i] = tsz.Point{math.MaxFloat64 / 100, ts}
		ConstantNearMinf[i] = tsz.Point{-math.MaxFloat64 / 100, ts}
		ConstantNearMax0f[i] = tsz.Point{math.Floor(ConstantNearMaxf[i].V), ts}
		ConstantNearMin0f[i] = tsz.Point{math.Floor(ConstantNearMinf[i].V), ts}
		SmallRangePosd[i] = tsz.Point{testdata.TwoHoursData[i%120].V, ts}
		LargeRangePosd[i] = tsz.Point{testdata.TwoHoursData[i%120].V * 1000000, ts}
		SmallRangePosf[i] = tsz.Point{float64(testdata.TwoHoursData[i%120].V) * 1.234567, ts}
		LargeRangePosf[i] = tsz.Point{float64(testdata.TwoHoursData[i%120].V) * 0.00001234567 * math.MaxFloat64, ts}
		SmallRangePos0f[i] = tsz.Point{math.Floor(SmallRangePosf[i].V), ts}
		LargeRangePos0f[i] = tsz.Point{math.Floor(LargeRangePosf[i].V), ts}
		RandomSmallPosf[i] = tsz.Point{rand.ExpFloat64(), ts}
		RandomSmallf[i] = tsz.Point{rand.NormFloat64(), ts}
		RandomLargePosf[i] = tsz.Point{rand.ExpFloat64() * 0.0001 * math.MaxFloat64, ts}
		RandomLargef[i] = tsz.Point{rand.NormFloat64() * 0.0001 * math.MaxFloat64, ts}
		RandomLargePos0f[i] = tsz.Point{math.Floor(RandomLargePosf[i].V), ts}
		RandomLarge0f[i] = tsz.Point{math.Floor(RandomLargef[i].V), ts}
	}

	intervals := []int{10, 30, 60, 120, 360, 720, 1440}
	do := func(data []tsz.Point) string {
		str := ""
		for _, points := range intervals {
			s := tsz.New(data[0].T)
			for _, tt := range data[0:points] {
				s.Push(tt.T, tt.V)
			}
			size := len(s.Bytes())
			BPerPoint := float64(size) / float64(points)
			str += fmt.Sprintf("\033[31m%d\033[39m\t%.2f\t", size, BPerPoint)
		}
		return str
	}
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 5, 0, 1, ' ', tabwriter.AlignRight)
	fmt.Fprintln(w, "=== help ===")
	fmt.Fprintln(w, "CS = chunk size in Bytes")
	fmt.Fprintln(w, "BPP = Bytes per point (CS/num-points)")
	fmt.Fprintln(w, "d = integers stored as float64")
	fmt.Fprintln(w, "f = float64's with a bunch of decimal numbers")
	fmt.Fprintln(w, ".Xf = float64's with X decimal numbers")
	fmt.Fprintln(w, "=== data ===")
	str := "test"
	for _, points := range intervals {
		str += fmt.Sprintf("\t  \033[39m%dCS\033[39m\t%dBPP", points, points)
	}
	fmt.Fprintln(w, str)
	fmt.Fprintf(w, "constant   zero            d\t"+do(ConstantZero)+"\n")
	fmt.Fprintf(w, "constant   one             d\t"+do(ConstantOne)+"\n")
	fmt.Fprintf(w, "constant   pos           .3f\t"+do(ConstantPos3f)+"\n")
	fmt.Fprintf(w, "constant   neg           .3f\t"+do(ConstantNeg3f)+"\n")
	fmt.Fprintf(w, "constant   pos           .0f\t"+do(ConstantPos0f)+"\n")
	fmt.Fprintf(w, "constant   neg           .0f\t"+do(ConstantNeg0f)+"\n")
	fmt.Fprintf(w, "constant   nearmax         f\t"+do(ConstantNearMaxf)+"\n")
	fmt.Fprintf(w, "constant   nearmin         f\t"+do(ConstantNearMinf)+"\n")
	fmt.Fprintf(w, "constant   nearmax       .0f\t"+do(ConstantNearMax0f)+"\n")
	fmt.Fprintf(w, "constant   nearmin       .0f\t"+do(ConstantNearMin0f)+"\n")
	fmt.Fprintf(w, "batch100   zero/one        d\t"+do(Batch100ZeroOne)+"\n")
	fmt.Fprintf(w, "flapping   zero/one        d\t"+do(FlappingZeroOne)+"\n")
	fmt.Fprintf(w, "pick-range small pos       d\t"+do(SmallRangePosd)+"\n")
	fmt.Fprintf(w, "pick-range large pos       d\t"+do(LargeRangePosd)+"\n")
	fmt.Fprintf(w, "pick-range small pos       f\t"+do(SmallRangePosf)+"\n")
	fmt.Fprintf(w, "pick-range large pos       f\t"+do(LargeRangePosf)+"\n")
	fmt.Fprintf(w, "pick-range small pos     .0f\t"+do(SmallRangePos0f)+"\n")
	fmt.Fprintf(w, "pick-range large pos     .0f\t"+do(LargeRangePos0f)+"\n")
	fmt.Fprintf(w, "pick-rand  small pos       f\t"+do(RandomSmallPosf)+"\n")
	fmt.Fprintf(w, "pick-rand  small pos/neg   f\t"+do(RandomSmallf)+"\n")
	fmt.Fprintf(w, "pick-rand  large pos       f\t"+do(RandomLargePosf)+"\n")
	fmt.Fprintf(w, "pick-rand  large pos/neg   f\t"+do(RandomLargef)+"\n")
	fmt.Fprintf(w, "pick-rand  large pos     .0f\t"+do(RandomLargePos0f)+"\n")
	fmt.Fprintf(w, "pick-rand  large pos/neg .0f\t"+do(RandomLarge0f)+"\n")
	fmt.Fprintln(w)
	w.Flush()
}
