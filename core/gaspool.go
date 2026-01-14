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

package core

import (
	"fmt"
	"math"

	"github.com/ethereum/go-ethereum/core/vm"
)

// GasPool tracks the amount of gas available during execution of the transactions
// in a block. The zero value is a pool with zero gas available.
type GasPool vm.GasCosts

// AddGas makes gas available for execution.
func (gp *GasPool) AddGas(amount uint64) *GasPool {
	if gp.RegularGas > math.MaxUint64-amount {
		panic("gas pool pushed above uint64")
	}
	gp.RegularGas += amount
	return gp
}

// SubGas deducts the given amount from the pool if enough gas is
// available and returns an error otherwise.
func (gp *GasPool) SubGas(amount uint64) error {
	if gp.RegularGas < amount {
		return ErrGasLimitReached
	}
	gp.RegularGas -= amount
	return nil
}

// Gas returns the amount of gas remaining in the pool.
func (gp *GasPool) Gas() uint64 {
	return gp.RegularGas
}

// SetGas sets the amount of gas with the provided number.
func (gp *GasPool) SetGas(gas uint64) {
	gp.RegularGas = gas
}

func (gp *GasPool) String() string {
	return fmt.Sprintf("<%d,%d>", gp.RegularGas, gp.StateGas)
}

func (gp *GasPool) SetDataGas(gas uint64) {
	gp.StateGas = gas
}
