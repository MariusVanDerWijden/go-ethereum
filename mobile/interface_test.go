// Copyright 2019 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package geth

import (
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

func TestInterfaceGetSet(t *testing.T) {
	var tests = []struct {
		method string
		input  interface{}
		expect interface{}
	}{
		{"Bool", true, true},
		{"Bool", false, false},
		{"Bools", &Bools{[]bool{false, true}}, &Bools{[]bool{false, true}}},
		{"String", "go-ethereum", "go-ethereum"},
		{"Strings", &Strings{strs: []string{"hello", "world"}}, &Strings{strs: []string{"hello", "world"}}},
		{"Binary", []byte{0x01, 0x02}, []byte{0x01, 0x02}},
		{"Binaries", &Binaries{[][]byte{{0x01, 0x02}, {0x03, 0x04}}}, &Binaries{[][]byte{{0x01, 0x02}, {0x03, 0x04}}}},
		{"Address", &Address{common.HexToAddress("deadbeef")}, &Address{common.HexToAddress("deadbeef")}},
		{"Addresses", &Addresses{[]common.Address{common.HexToAddress("deadbeef"), common.HexToAddress("cafebabe")}}, &Addresses{[]common.Address{common.HexToAddress("deadbeef"), common.HexToAddress("cafebabe")}}},
		{"Hash", &Hash{common.HexToHash("deadbeef")}, &Hash{common.HexToHash("deadbeef")}},
		{"Hashes", &Hashes{[]common.Hash{common.HexToHash("deadbeef"), common.HexToHash("cafebabe")}}, &Hashes{[]common.Hash{common.HexToHash("deadbeef"), common.HexToHash("cafebabe")}}},
		{"Int8", int8(1), int8(1)},
		{"Int16", int16(1), int16(1)},
		{"Int32", int32(1), int32(1)},
		{"Int64", int64(1), int64(1)},
		{"Int8s", &BigInts{[]*big.Int{big.NewInt(1), big.NewInt(2)}}, &BigInts{[]*big.Int{big.NewInt(1), big.NewInt(2)}}},
		{"Int16s", &BigInts{[]*big.Int{big.NewInt(1), big.NewInt(2)}}, &BigInts{[]*big.Int{big.NewInt(1), big.NewInt(2)}}},
		{"Int32s", &BigInts{[]*big.Int{big.NewInt(1), big.NewInt(2)}}, &BigInts{[]*big.Int{big.NewInt(1), big.NewInt(2)}}},
		{"Int64s", &BigInts{[]*big.Int{big.NewInt(1), big.NewInt(2)}}, &BigInts{[]*big.Int{big.NewInt(1), big.NewInt(2)}}},
		{"Uint8", NewBigInt(1), NewBigInt(1)},
		{"Uint16", NewBigInt(1), NewBigInt(1)},
		{"Uint32", NewBigInt(1), NewBigInt(1)},
		{"Uint64", NewBigInt(1), NewBigInt(1)},
		{"Uint8s", &BigInts{[]*big.Int{big.NewInt(1), big.NewInt(2)}}, &BigInts{[]*big.Int{big.NewInt(1), big.NewInt(2)}}},
		{"Uint16s", &BigInts{[]*big.Int{big.NewInt(1), big.NewInt(2)}}, &BigInts{[]*big.Int{big.NewInt(1), big.NewInt(2)}}},
		{"Uint32s", &BigInts{[]*big.Int{big.NewInt(1), big.NewInt(2)}}, &BigInts{[]*big.Int{big.NewInt(1), big.NewInt(2)}}},
		{"Uint64s", &BigInts{[]*big.Int{big.NewInt(1), big.NewInt(2)}}, &BigInts{[]*big.Int{big.NewInt(1), big.NewInt(2)}}},
		{"BigInt", NewBigInt(1), NewBigInt(1)},
		{"BigInts", &BigInts{[]*big.Int{big.NewInt(1), big.NewInt(2)}}, &BigInts{[]*big.Int{big.NewInt(1), big.NewInt(2)}}},
	}

	args := NewInterfaces(len(tests))

	callFn := func(receiver interface{}, method string, arg interface{}) interface{} {
		rval := reflect.ValueOf(receiver)
		rval.MethodByName(fmt.Sprintf("Set%s", method)).Call([]reflect.Value{reflect.ValueOf(arg)})
		res := rval.MethodByName(fmt.Sprintf("Get%s", method)).Call(nil)
		if len(res) > 0 {
			return res[0].Interface()
		}
		return nil
	}

	for index, c := range tests {
		// In theory the change of iface shouldn't effect the args value
		iface, _ := args.Get(index)
		result := callFn(iface, c.method, c.input)
		if !reflect.DeepEqual(result, c.expect) {
			t.Errorf("Interface get/set mismatch, want %v, got %v", c.expect, result)
		}
		// Check whether the underlying value in args is still zero
		iface, _ = args.Get(index)
		if iface.object != nil {
			t.Error("Get operation is not write safe")
		}
	}
}

func TestInterfacesFromJSON(t *testing.T) {
	args := NewInterfaces(2)
	args1 := NewInterface()
	args1.SetBigInt(&BigInt{big.NewInt(1)})
	args.Set(0, args1)
	args2 := NewInterface()
	args2.SetBigInt(&BigInt{big.NewInt(2)})
	args.Set(1, args2)

	definition := "{\"inputs\" :[{ \"indexed\":false, \"name\":\"t\", \"type\":\"tuple\", \"components\":[{\"name\":\"x\", \"type\":\"uint256\"}, {\"name\":\"y\", \"type\":\"uint256\" }]}]}"
	interf, err := NewInterfacesFromJSON(definition, args)
	if err != nil {
		t.Error(err)
	}

	jsondef := "[{\"anonymous\":false,\"inputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"a\",\"type\":\"uint256\"},{\"internalType\":\"uint256[]\",\"name\":\"b\",\"type\":\"uint256[]\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"X\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"y\",\"type\":\"uint256\"}],\"internalType\":\"structT[]\",\"name\":\"c\",\"type\":\"tuple[]\"}],\"indexed\":false,\"internalType\":\"structTupleTest2.S\",\"name\":\"\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"X\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"Y\",\"type\":\"uint256\"}],\"indexed\":false,\"internalType\":\"structT\",\"name\":\"\",\"type\":\"tuple\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"evF\",\"type\":\"event\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"X\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"y\",\"type\":\"uint256\"}],\"internalType\":\"structT\",\"name\":\"t\",\"type\":\"tuple\"}],\"name\":\"a\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"X\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"y\",\"type\":\"uint256\"}],\"internalType\":\"structT\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"X\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"y\",\"type\":\"uint256\"}],\"internalType\":\"structT[]\",\"name\":\"t\",\"type\":\"tuple[]\"}],\"name\":\"b\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"X\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"y\",\"type\":\"uint256\"}],\"internalType\":\"structT[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"a\",\"type\":\"uint256\"},{\"internalType\":\"uint256[]\",\"name\":\"b\",\"type\":\"uint256[]\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"X\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"y\",\"type\":\"uint256\"}],\"internalType\":\"structT[]\",\"name\":\"c\",\"type\":\"tuple[]\"}],\"internalType\":\"structTupleTest2.S\",\"name\":\"s\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"X\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"y\",\"type\":\"uint256\"}],\"internalType\":\"structT\",\"name\":\"t\",\"type\":\"tuple\"},{\"internalType\":\"uint256\",\"name\":\"u\",\"type\":\"uint256\"}],\"name\":\"f\",\"outputs\":[],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"g\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"a\",\"type\":\"uint256\"},{\"internalType\":\"uint256[]\",\"name\":\"b\",\"type\":\"uint256[]\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"X\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"y\",\"type\":\"uint256\"}],\"internalType\":\"structT[]\",\"name\":\"c\",\"type\":\"tuple[]\"}],\"internalType\":\"structTupleTest2.S\",\"name\":\"\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"X\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"y\",\"type\":\"uint256\"}],\"internalType\":\"structT\",\"name\":\"\",\"type\":\"tuple\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"a\",\"type\":\"uint256\"},{\"internalType\":\"uint256[]\",\"name\":\"b\",\"type\":\"uint256[]\"}],\"internalType\":\"structTupleTest2.A\",\"name\":\"a\",\"type\":\"tuple\"},{\"internalType\":\"uint256\",\"name\":\"b\",\"type\":\"uint256\"}],\"name\":\"method\",\"outputs\":[],\"stateMutability\":\"view\",\"type\":\"function\"}]"
	exp, err := abi.JSON(strings.NewReader(jsondef))
	if err != nil {
		t.Error(err)
	}

	packed, err := exp.Pack("a", interf.object)
	if err != nil {
		t.Error(err)
	}

	defInterf, err := NewInterfacesFromJSONDefault(definition)
	if err != nil {
		t.Error(err)
	}

	err = exp.Unpack(defInterf.object, "a", packed[4:])
	if err != nil {
		t.Error(err)
	}
}
