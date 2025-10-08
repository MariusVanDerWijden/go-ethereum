// Copyright 2022 The go-ethereum Authors
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

package native

import (
	"encoding/json"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/eth/tracers"
	"github.com/ethereum/go-ethereum/params"
)

func init() {
	tracers.DefaultDirectory.Register("opCountTracer", newOpCountTracer, false)
}

type opCountTracer struct {
	err       error
	interrupt atomic.Bool
	opcodes   []map[byte]int
}

func newOpCountTracer(ctx *tracers.Context, cfg json.RawMessage, chainConfig *params.ChainConfig) (*tracers.Tracer, error) {
	t := new(opCountTracer)
	t.opcodes = make([]map[byte]int, 0)
	return &tracers.Tracer{
		Hooks: &tracing.Hooks{
			OnTxStart: t.OnTxStart,
			OnOpcode:  t.OnOpcode,
		},
		GetResult: t.GetResult,
		Stop:      t.Stop,
	}, nil
}

// OnOpcode implements the EVMLogger interface to trace a single step of VM execution.
func (t *opCountTracer) OnOpcode(pc uint64, opcode byte, gas, cost uint64, scope tracing.OpContext, rData []byte, depth int, err error) {
	if err != nil {
		return
	}
	// Skip if tracing was interrupted
	if t.interrupt.Load() {
		return
	}
	op := vm.OpCode(opcode)
	t.opcodes[len(t.opcodes)-1][byte(op)]++
}

func (t *opCountTracer) OnTxStart(env *tracing.VMContext, tx *types.Transaction, from common.Address) {
	t.opcodes = append(t.opcodes, make(map[byte]int))
}

// GetResult returns the json-encoded nested list of call traces, and any
// error arising from the encoding or forceful termination (via `Stop`).
func (t *opCountTracer) GetResult() (json.RawMessage, error) {
	var res []byte
	for m := range t.opcodes {
		msg, err := json.Marshal(m)
		if err != nil {
			return nil, err
		}
		res = append(res, msg...)
	}
	return json.RawMessage(res), t.err
}

// Stop terminates execution of the tracer at the first opportune moment.
func (t *opCountTracer) Stop(err error) {
	t.err = err
	t.interrupt.Store(true)
}
