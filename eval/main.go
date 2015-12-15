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
var DataZeroFloats = make([]testdata.Point, 60*24)
var DataSameFloats = make([]testdata.Point, 60*24)
var DataSmallRangePosInts = make([]testdata.Point, 60*24)
var DataSmallRangePosFloats = make([]testdata.Point, 60*24)
var DataLargeRangePosFloats = make([]testdata.Point, 60*24)
var DataRandomPosFloats = make([]testdata.Point, 60*24)
var DataRandomFloats = make([]testdata.Point, 60*24)

func init() {
	for i := 0; i < len(DataZeroFloats); i++ {
		DataZeroFloats[i] = testdata.Point{float64(0), uint32(i * 60)}
	}
	for i := 0; i < len(DataSameFloats); i++ {
		DataSameFloats[i] = testdata.Point{float64(1234.567), uint32(i * 60)}
	}
	for i := 0; i < len(DataSmallRangePosInts); i++ {
		DataSmallRangePosInts[i] = testdata.Point{testdata.TwoHoursData[i%120].V, uint32(i * 60)}
	}
	for i := 0; i < len(DataSmallRangePosFloats); i++ {
		v := float64(testdata.TwoHoursData[i%120].V) * 1.2
		DataSmallRangePosFloats[i] = testdata.Point{v, uint32(i * 60)}
	}
	for i := 0; i < len(DataLargeRangePosFloats); i++ {
		v := (float64(testdata.TwoHoursData[i%120].V) / 1000) * math.MaxFloat64
		DataLargeRangePosFloats[i] = testdata.Point{v, uint32(i * 60)}
	}
	for i := 0; i < len(DataRandomPosFloats); i++ {
		DataRandomPosFloats[i] = testdata.Point{rand.ExpFloat64(), uint32(i * 60)}
	}
	for i := 0; i < len(DataRandomFloats); i++ {
		DataRandomFloats[i] = testdata.Point{rand.NormFloat64(), uint32(i * 60)}
	}
}
func main() {
	intervals := []int{10, 30, 60, 120, 360, 720, 1440}
	do := func(data []testdata.Point) string {
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
	fmt.Fprintln(w, "CS = chunk size in Bytes")
	fmt.Fprintln(w, "BPP = Bytes per point")
	str := "test"
	for _, points := range intervals {
		str += fmt.Sprintf("\t  \033[39m%dCS\033[39m\t%dBPP", points, points)
	}
	fmt.Fprintln(w, str)
	fmt.Fprintf(w, "all zeroes      float\t"+do(DataZeroFloats)+"\n")
	fmt.Fprintf(w, "all the same    float\t"+do(DataSameFloats)+"\n")
	fmt.Fprintf(w, "small range pos   int\t"+do(DataSmallRangePosInts)+"\n")
	fmt.Fprintf(w, "small range pos float\t"+do(DataSmallRangePosFloats)+"\n")
	fmt.Fprintf(w, "large range pos float\t"+do(DataLargeRangePosFloats)+"\n")
	fmt.Fprintf(w, "random positive float\t"+do(DataRandomPosFloats)+"\n")
	fmt.Fprintf(w, "random pos/neg  float\t"+do(DataRandomFloats)+"\n")
	fmt.Fprintln(w)
	w.Flush()
}
