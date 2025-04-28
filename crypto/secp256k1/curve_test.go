package secp256k1

import (
	"math/big"
	"testing"
)

func TestAffine(t *testing.T) {
	x := new(big.Int)
	y := new(big.Int)
	z := new(big.Int)
	theCurve.affineFromJacobian(x, y, z)
}

func TestAdd(t *testing.T) {
	x := new(big.Int)
	y := new(big.Int)
	z := new(big.Int)
	w := new(big.Int)
	theCurve.Add(x, y, z, w)
}

func TestDouble(t *testing.T) {
	x := new(big.Int)
	y := new(big.Int)
	theCurve.Double(x, y)
}
