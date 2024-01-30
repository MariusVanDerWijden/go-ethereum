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
}

func strToFe(s string) *fe {
	x, ok := new(big.Int).SetString(s, 10)
	if !ok {
		panic("not okay")
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
