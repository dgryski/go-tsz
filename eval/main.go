package main

import (
	"fmt"
	"github.com/dgryski/go-tsz"
	"math"
	"math/rand"
)

// collection of 24h worth of minutely points, with different characteristics.
var DataZeroFloats = make([]tsz.Point, 60*24)
var DataSameFloats = make([]tsz.Point, 60*24)
var DataSmallRangePosInts = make([]tsz.Point, 60*24)
var DataSmallRangePosFloats = make([]tsz.Point, 60*24)
var DataLargeRangePosFloats = make([]tsz.Point, 60*24)
var DataRandomPosFloats = make([]tsz.Point, 60*24)
var DataRandomFloats = make([]tsz.Point, 60*24)

func init() {
	for i := 0; i < len(DataZeroFloats); i++ {
		DataZeroFloats[i] = tsz.Point{float64(0), uint32(i * 60)}
	}
	for i := 0; i < len(DataSameFloats); i++ {
		DataSameFloats[i] = tsz.Point{float64(1234.567), uint32(i * 60)}
	}
	for i := 0; i < len(DataSmallRangePosInts); i++ {
		DataSmallRangePosInts[i] = tsz.Point{tsz.TwoHoursData[i%120].V, uint32(i * 60)}
	}
	for i := 0; i < len(DataSmallRangePosFloats); i++ {
		v := float64(tsz.TwoHoursData[i%120].V) * 1.2
		DataSmallRangePosFloats[i] = tsz.Point{v, uint32(i * 60)}
	}
	for i := 0; i < len(DataLargeRangePosFloats); i++ {
		v := (float64(tsz.TwoHoursData[i%120].V) / 1000) * math.MaxFloat64
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
	benchmarkEncodeSize(10)
	benchmarkEncodeSize(30)
	benchmarkEncodeSize(60)
	benchmarkEncodeSize(120)
	benchmarkEncodeSize(360)
	benchmarkEncodeSize(720)
	benchmarkEncodeSize(1440)
}

func benchmarkEncodeSize(points int) {
	do := func(data []tsz.Point, desc string) {
		s := tsz.New(data[0].T)
		for _, tt := range data[0:points] {
			s.Push(tt.T, tt.V)
		}
		size := len(s.Bytes())
		BPerPoint := float64(size) / float64(points)
		fmt.Printf("encode %d %s points: %6d Bytes. %.2f B/point\n", points, desc, size, BPerPoint)
	}
	do(DataZeroFloats, "all zeroes      float")
	do(DataSameFloats, "all the same    float")
	do(DataSmallRangePosInts, "small range pos   int")
	do(DataSmallRangePosFloats, "small range pos float")
	do(DataLargeRangePosFloats, "large range pos float")
	do(DataRandomPosFloats, "random positive float")
	do(DataRandomFloats, "random pos/neg  float")
}
