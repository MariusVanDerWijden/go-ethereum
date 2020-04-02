// Copyright 2015 The go-ethereum Authors
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

package abi

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
)

// Method represents a callable given a `Name` and whether the method is a constant.
// If the method is `Const` no transaction needs to be created for this
// particular Method call. It can easily be simulated using a local VM.
// For example a `Balance()` method only needs to retrieve something
// from the storage and therefore requires no Tx to be send to the
// network. A method such as `Transact` does require a Tx and thus will
// be flagged `false`.
// Input specifies the required input parameters for this gives method.
type Method struct {
	// Name is the method name used for internal representation. It's derived from
	// the raw name and a suffix will be added in the case of a function overload.
	//
	// e.g.
	// There are two functions have same name:
	// * foo(int,int)
	// * foo(uint,uint)
	// The method name of the first one will be resolved as foo while the second one
	// will be resolved as foo0.
	Name string
	// RawName is the raw method name parsed from ABI.
	RawName string
	Const   bool
	Inputs  Arguments
	Outputs Arguments
	str     string
	// Sig returns the methods string signature according to the ABI spec.
	// e.g.		function foo(uint32 a, int b) = "foo(uint32,int256)"
	// Please note that "int" is substitute for its canonical representation "int256"
	Sig string
	// ID returns the canonical representation of the method's signature used by the
	// abi definition to identify method names and types.
	ID []byte
}

// NewMethod creates a new Method.
// It also precomputes the sig representation and the string representation
// of the method.
// A method should always be created using NewMethod.
func NewMethod(name string, rawName string, isConst bool, inputs Arguments, outputs Arguments) Method {
	// inputs
	inputNames := make([]string, len(inputs))
	types := make([]string, len(inputs))
	for i, input := range inputs {
		inputNames[i] = fmt.Sprintf("%v %v", input.Type, input.Name)
		types[i] = input.Type.String()
	}
	// outputs
	outputNames := make([]string, len(outputs))
	for i, output := range outputs {
		outputNames[i] = output.Type.String()
		if len(output.Name) > 0 {
			outputNames[i] += fmt.Sprintf(" %v", output.Name)
		}
	}
	constant := ""
	if isConst {
		constant = "constant "
	}

	str := fmt.Sprintf("function %v(%v) %sreturns(%v)", rawName, strings.Join(inputNames, ", "), constant, strings.Join(outputNames, ", "))
	sig := fmt.Sprintf("%v(%v)", rawName, strings.Join(types, ","))
	id := crypto.Keccak256([]byte(sig))[:4]

	method := Method{
		Name:    name,
		RawName: rawName,
		Const:   isConst,
		Inputs:  inputs,
		Outputs: outputs,
		str:     str,
		Sig:     sig,
		ID:      id,
	}
	return method
}

func (method Method) String() string {
	return method.str
}
