package bls12381

import (
	"fmt"
	"math/big"
	"testing"
)

func TestParams(t *testing.T) {
	x1 := "1586958781458431025242759403266842894121773480562120986020912974854563298150952611241517463240701"
	fmt.Printf("%v\n", strToFe(x1))
	x2 := "2001204777610833696708894912867952078278441409969503942666029068062015825245418932221343814564507832018947136279894"
	fmt.Printf("%v\n", strToFe(x2))
	b := "0x4"
	fmt.Printf("%v\n", strToFe(b))
}

func TestSwiftEC(tt *testing.T) {
	u := "0x188114530a157117accadf5223257941ab6148fc30a0a75a143df1b86b023ac2ea1b403465cefd84b6da5de70db194f3"
	t := "0x9135b44e251b5562b953fdaef68fc1043eb0ace10982588b0d21479d8107d3fec503fa4ef6b78397b3c4aef331f688b"
	s := "0"
	xx, yy := ecMapG1(strToFe(u), strToFe(t), strToFe(s))
	x := "0x14c8a8e700758a0ba0a42073b988fc4fcd26d714289fe43f8a07a346db61b87abf39653cea9bada9c9bf69be3be08670"
	y := "0x0ee5380b7f6733867137bed2aae36aa92823685b3f08f38e91ac4472ba192a7cd289f4146d84a997cc334715be7d53ae"

	if toString(xx) != x {
		tt.Errorf("expected different encoding: %v got %v", toString(xx), x)
	}
	if toString(yy) != y {
		tt.Errorf("expected different encoding: %v got %v", toString(yy), y)
	}
}

func strToFe(s string) *fe {
	x, ok := new(big.Int).SetString(s, 0)
	if !ok {
		panic(fmt.Sprintf("not okay: %v", s))
	}
	txt := x.Text(16)
	if len(txt)%2 == 1 {
		txt = "0" + txt
	}
	x11, err := fromString(txt)
	if err != nil {
		panic(err)
	}

	return x11
}

func BenchmarkECMapG1(b *testing.B) {
	u := "0x188114530a157117accadf5223257941ab6148fc30a0a75a143df1b86b023ac2ea1b403465cefd84b6da5de70db194f3"
	t := "0x9135b44e251b5562b953fdaef68fc1043eb0ace10982588b0d21479d8107d3fec503fa4ef6b78397b3c4aef331f688b"
	s := "0"
	ui, ti, si := strToFe(u), strToFe(t), strToFe(s)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ecMapG1(ui, ti, si)
	}
}

func BenchmarkSwu(b *testing.B) {
	u := "0x188114530a157117accadf5223257941ab6148fc30a0a75a143df1b86b023ac2ea1b403465cefd84b6da5de70db194f3"
	t := "0x9135b44e251b5562b953fdaef68fc1043eb0ace10982588b0d21479d8107d3fec503fa4ef6b78397b3c4aef331f688b"
	s := "0"
	ui, _, _ := strToFe(u), strToFe(t), strToFe(s)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x, y := swuMapG1(ui)
		isogenyMapG1(x, y)
	}
}

func BenchmarkECMapG2(b *testing.B) {
	params := swuParamsForG2
	for i := 0; i < b.N; i++ {
		ecMapG2(params.a, params.a, params.a)
	}
}

func BenchmarkSWUG2(b *testing.B) {
	params := swuParamsForG2
	for i := 0; i < b.N; i++ {
		x, y := swuMapG2(nil, params.a)
		isogenyMapG2(nil, x, y)
	}
}

// BenchmarkSwu-24    	   18744	     66977 ns/op	     288 B/op	       6 allocs/op
// BenchmarkECMapG1-24    	    8391	    139731 ns/op	     288 B/op	       6 allocs/op
// BenchmarkECMapG1-24    	   18232	     67647 ns/op	      96 B/op	       2 allocs/op
// BenchmarkSwu-24    	   15289	     78494 ns/op	     288 B/op	       6 allocs/op
