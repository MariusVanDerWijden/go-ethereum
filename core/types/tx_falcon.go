// Copyright 2021 The go-ethereum Authors
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

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
)

// FalconTx represents an RIP-??? transaction.
type FalconTx struct {
	ChainID    *big.Int
	Nonce      uint64
	GasTipCap  *big.Int // a.k.a. maxPriorityFeePerGas
	GasFeeCap  *big.Int // a.k.a. maxFeePerGas
	Gas        uint64
	To         *common.Address `rlp:"nil"` // nil means contract creation
	Value      *big.Int
	Data       []byte
	AccessList AccessList

	// Signature values
	Sender    common.Address
	Signature []byte
}

// copy creates a deep copy of the transaction data and initializes all fields.
func (tx *FalconTx) copy() TxData {
	cpy := &FalconTx{
		Nonce: tx.Nonce,
		To:    copyAddressPtr(tx.To),
		Data:  common.CopyBytes(tx.Data),
		Gas:   tx.Gas,
		// These are copied below.
		AccessList: make(AccessList, len(tx.AccessList)),
		Value:      new(big.Int),
		ChainID:    new(big.Int),
		GasTipCap:  new(big.Int),
		GasFeeCap:  new(big.Int),
		Signature:  common.CopyBytes(tx.Signature),
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
	return cpy
}

// accessors for innerTx.
func (tx *FalconTx) txType() byte           { return FalconTxType }
func (tx *FalconTx) chainID() *big.Int      { return tx.ChainID }
func (tx *FalconTx) accessList() AccessList { return tx.AccessList }
func (tx *FalconTx) data() []byte           { return tx.Data }
func (tx *FalconTx) gas() uint64            { return tx.Gas }
func (tx *FalconTx) gasFeeCap() *big.Int    { return tx.GasFeeCap }
func (tx *FalconTx) gasTipCap() *big.Int    { return tx.GasTipCap }
func (tx *FalconTx) gasPrice() *big.Int     { return tx.GasFeeCap }
func (tx *FalconTx) value() *big.Int        { return tx.Value }
func (tx *FalconTx) nonce() uint64          { return tx.Nonce }
func (tx *FalconTx) to() *common.Address    { return tx.To }

func (tx *FalconTx) effectiveGasPrice(dst *big.Int, baseFee *big.Int) *big.Int {
	if baseFee == nil {
		return dst.Set(tx.GasFeeCap)
	}
	tip := dst.Sub(tx.GasFeeCap, baseFee)
	if tip.Cmp(tx.GasTipCap) > 0 {
		tip.Set(tx.GasTipCap)
	}
	return tip.Add(tip, baseFee)
}

func (tx *FalconTx) rawSignatureValues() (v, r, s *big.Int) {
	return nil, nil, nil
}

func (tx *FalconTx) setSignatureValues(chainID, v, r, s *big.Int) {
	tx.ChainID = chainID
}

func (tx *FalconTx) encode(b *bytes.Buffer) error {
	return rlp.Encode(b, tx)
}

func (tx *FalconTx) decode(input []byte) error {
	return rlp.DecodeBytes(input, tx)
}
