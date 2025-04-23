// Copyright 2025 The go-ethereum Authors
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

package types

import (
	"bytes"
	"math/big"
	"slices"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/holiman/uint256"
)

// CreateTx represents an EIP-7873 transaction.
type CreateTx struct {
	ChainID    *uint256.Int
	Nonce      uint64
	GasTipCap  *uint256.Int // a.k.a. maxPriorityFeePerGas
	GasFeeCap  *uint256.Int // a.k.a. maxFeePerGas
	Gas        uint64
	To         common.Address
	Value      *uint256.Int
	Data       []byte
	AccessList AccessList
	Initcodes  [][]byte

	// Signature values
	V *uint256.Int
	R *uint256.Int
	S *uint256.Int
}

// copy creates a deep copy of the transaction data and initializes all fields.
func (tx *CreateTx) copy() TxData {
	cpy := &CreateTx{
		Nonce: tx.Nonce,
		To:    tx.To,
		Data:  common.CopyBytes(tx.Data),
		Gas:   tx.Gas,
		// These are copied below.
		AccessList: make(AccessList, len(tx.AccessList)),
		Initcodes:  make([][]byte, len(tx.Initcodes)),
		Value:      new(uint256.Int),
		ChainID:    new(uint256.Int),
		GasTipCap:  new(uint256.Int),
		GasFeeCap:  new(uint256.Int),
		V:          new(uint256.Int),
		R:          new(uint256.Int),
		S:          new(uint256.Int),
	}
	copy(cpy.AccessList, tx.AccessList)
	if tx.Value != nil {
		cpy.Value.Set(tx.Value)
	}
	if tx.ChainID != nil {
		cpy.ChainID.Set(tx.ChainID)
	}
	if tx.GasTipCap != nil {
		cpy.GasTipCap.Set(tx.GasTipCap)
	}
	if tx.GasFeeCap != nil {
		cpy.GasFeeCap.Set(tx.GasFeeCap)
	}
	if tx.V != nil {
		cpy.V.Set(tx.V)
	}
	if tx.R != nil {
		cpy.R.Set(tx.R)
	}
	if tx.S != nil {
		cpy.S.Set(tx.S)
	}
	if tx.Initcodes != nil {
		for _, inner := range tx.Initcodes {
			tx.Initcodes = append(cpy.Initcodes, slices.Clone(inner))
		}
	}
	return cpy
}

// accessors for innerTx.
func (tx *CreateTx) txType() byte           { return CreateTxType }
func (tx *CreateTx) chainID() *big.Int      { return tx.ChainID.ToBig() }
func (tx *CreateTx) accessList() AccessList { return tx.AccessList }
func (tx *CreateTx) data() []byte           { return tx.Data }
func (tx *CreateTx) gas() uint64            { return tx.Gas }
func (tx *CreateTx) gasFeeCap() *big.Int    { return tx.GasFeeCap.ToBig() }
func (tx *CreateTx) gasTipCap() *big.Int    { return tx.GasTipCap.ToBig() }
func (tx *CreateTx) gasPrice() *big.Int     { return tx.GasFeeCap.ToBig() }
func (tx *CreateTx) value() *big.Int        { return tx.Value.ToBig() }
func (tx *CreateTx) nonce() uint64          { return tx.Nonce }
func (tx *CreateTx) to() *common.Address    { tmp := tx.To; return &tmp }

func (tx *CreateTx) effectiveGasPrice(dst *big.Int, baseFee *big.Int) *big.Int {
	if baseFee == nil {
		return dst.Set(tx.GasFeeCap.ToBig())
	}
	tip := dst.Sub(tx.GasFeeCap.ToBig(), baseFee)
	if tip.Cmp(tx.GasTipCap.ToBig()) > 0 {
		tip.Set(tx.GasTipCap.ToBig())
	}
	return tip.Add(tip, baseFee)
}

func (tx *CreateTx) rawSignatureValues() (v, r, s *big.Int) {
	return tx.V.ToBig(), tx.R.ToBig(), tx.S.ToBig()
}

func (tx *CreateTx) setSignatureValues(chainID, v, r, s *big.Int) {
	tx.ChainID = uint256.MustFromBig(chainID)
	tx.V.SetFromBig(v)
	tx.R.SetFromBig(r)
	tx.S.SetFromBig(s)
}

func (tx *CreateTx) encode(b *bytes.Buffer) error {
	return rlp.Encode(b, tx)
}

func (tx *CreateTx) decode(input []byte) error {
	return rlp.DecodeBytes(input, tx)
}

func (tx *CreateTx) sigHash(chainID *big.Int) common.Hash {
	return prefixedRlpHash(
		CreateTxType,
		[]any{
			chainID,
			tx.Nonce,
			tx.GasTipCap,
			tx.GasFeeCap,
			tx.Gas,
			tx.To,
			tx.Value,
			tx.Data,
			tx.AccessList,
			tx.Initcodes,
		})
}
