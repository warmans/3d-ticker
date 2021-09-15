package main

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

var sampleAPIResponse = `{
  "prices": [
    [
      1631624546813,
      29.634300756946963
    ],
    [
      1631628534330,
      28.529577037197395
    ],
    [
      1631632031401,
      29.166992048863158
    ],
    [
      1631635654154,
      28.829803419737566
    ],
    [
      1631639317258,
      29.09728755391682
    ],
    [
      1631642829717,
      29.026176372436257
    ],
    [
      1631646582169,
      28.281487384022284
    ],
    [
      1631649807855,
      28.680200081280457
    ],
    [
      1631653289531,
      29.197292223796396
    ],
    [
      1631656889109,
      28.67916790037302
    ],
    [
      1631660597159,
      28.906579450328508
    ],
    [
      1631664263002,
      29.090397296105206
    ],
    [
      1631667763211,
      28.65979157147275
    ],
    [
      1631671688461,
      28.18814832912057
    ],
    [
      1631675039969,
      28.191369200576503
    ],
    [
      1631679150406,
      27.961417361925704
    ],
    [
      1631682276471,
      28.212557908723376
    ],
    [
      1631685809867,
      28.447983231696316
    ],
    [
      1631689371704,
      28.428419233349526
    ],
    [
      1631692978951,
      28.888010800093365
    ],
    [
      1631696546252,
      28.71228645758671
    ],
    [
      1631700507557,
      28.34986272388235
    ],
    [
      1631703711556,
      28.36524941584422
    ],
    [
      1631707356645,
      28.165867060814424
    ],
    [
      1631710624000,
      28.195390435908603
    ]
  ]
}`

func musDecodeAPIResponse(t *testing.T) *PriceData {
	prices := &PriceData{}
	err := json.NewDecoder(strings.NewReader(sampleAPIResponse)).Decode(&prices)
	require.NoError(t, err)
	return prices
}

func testVals(prices [][]float64) []float64 {
	vals := make([]float64, len(prices))
	for k, v := range prices {
		vals[k] = v[1]
	}
	return vals
}

func TestScalePrice(t *testing.T) {
	data := musDecodeAPIResponse(t)
	require.EqualValues(t, uint8(0), ScalePrice(testVals(data.Prices), 28))
	require.EqualValues(t, uint8(5), ScalePrice(testVals(data.Prices), 29))
}

func TestGroupSeries_BestCase(t *testing.T) {
	values := [][]float64{
		{0, 1.0},
		{0, 2.0},
		{0, 3.0},
		{0, 4.0},
		{0, 5.0},
		{0, 6.0},
		{0, 7.0},
		{0, 8.0},
		{0, 9.0},
		{0, 10.0},
	}
	converted := GroupSeries(values)
	require.EqualValues(t, []float64{1.5, 3.5, 5.5, 7.5, 9.5}, converted)
}

func TestFormatDataForDisplay_BestCase(t *testing.T) {
	values := [][]float64{
		{0, 1.0},
		{0, 2.0},
		{0, 3.0},
		{0, 4.0},
		{0, 5.0},
		{0, 6.0},
		{0, 7.0},
		{0, 8.0},
		{0, 9.0},
		{0, 10.0},
	}
	converted := FormatDataForDisplay(values)
	require.EqualValues(t, []uint8{0, 2, 4, 6, 9}, converted)
}

func TestFormatDataForDisplay_Realistic(t *testing.T) {
	data := musDecodeAPIResponse(t)
	converted := FormatDataForDisplay(data.Prices)
	require.EqualValues(t, 5, len(converted))
	fmt.Println(converted)
}
